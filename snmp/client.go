package snmp

import (
	"fmt"
	wapsnmp "github.com/cdevr/WapSNMP"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	PORT    = "161"
	TIMEOUT = time.Duration(2) * time.Second
	RETRIES = 3
	VERSION = wapsnmp.SNMPv2c
)

type Client struct {
	log *log.Logger

	version   wapsnmp.SNMPVersion
	community string
	timeout   time.Duration
	retries   int

	udpMutex sync.Mutex
	udpConn  *net.UDPConn
	udpSize  uint

	requestMutex sync.Mutex
	requestId    int32
	requests     map[int32]*request
}

func (self Client) String() string {
	return fmt.Sprintf("%s", self.udpConn.RemoteAddr())
}

type request struct {
	packet Packet
	pdu    PDU

	responseChan chan PDU
}

func (self request) String() string {
	return fmt.Sprintf("%d", self.pdu.RequestID)
}

type RequestError struct {
	RequestType wapsnmp.BERType
	ErrorStatus int
	ErrorIndex  int
}

func (self RequestError) Error() string {
	return fmt.Sprintf("SNMP error:%d [%d]", self.ErrorStatus, self.ErrorIndex)
}

func Connect(config Config) (*Client, error) {
	client := &Client{
		log: log.New(ioutil.Discard, "", 0),

		version:   VERSION,
		community: config.Community,
		timeout:   TIMEOUT,
		retries:   RETRIES,

		udpSize: UDP_SIZE,

		requestId: 1337,
		requests:  make(map[int32]*request),
	}

	if err := client.connect(config); err != nil {
		return nil, err
	}

	// process responses
	go client.recvLoop()

	return client, nil
}

func (self *Client) connect(config Config) error {
	port := config.Port

	if port == "" {
		port = PORT
	}

	netAddr := net.JoinHostPort(config.Host, port)

	if udpAddr, err := net.ResolveUDPAddr("udp", netAddr); err != nil {
		return err
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		return err
	} else {
		self.udpConn = udpConn
	}

	return nil
}

func (self *Client) Log() {
	self.log = log.New(os.Stderr, fmt.Sprintf("snmp.Client %v: ", self), 0)
}

// Encode and send a generic PDU
func (self *Client) sendPDU(packet Packet, pdu PDU) error {
	self.udpMutex.Lock()
	defer self.udpMutex.Unlock()

	// send
	deadline := time.Now().Add(self.timeout)

	if err := self.udpConn.SetWriteDeadline(deadline); err != nil {
		return err
	}

	if buf, err := encodePacket(packet, packPDU(packet.PduType, pdu)); err != nil {
		return err
	} else if size, err := self.udpConn.Write(buf); err != nil {
		return err
	} else if size != len(buf) {
		return fmt.Errorf("short write: %d < %d", size, len(buf))
	}

	return nil
}

// Recv and decode a generic PDU
func (self *Client) recv() (packet Packet, pdu PDU, err error) {
	// recv
	buf := make([]byte, self.udpSize)

	size, err := self.udpConn.Read(buf)

	if err != nil {
		return packet, pdu, err
	} else if size == 0 {
		return packet, pdu, fmt.Errorf("EOF")
	}

	// parse
	if packet, packetPdu, err := decodePacket(buf[:size]); err != nil {
		return packet, pdu, fmt.Errorf("invalid packet: %s", err)
	} else if pdu, err := unpackPDU(packetPdu); err != nil {
		return packet, pdu, fmt.Errorf("invalid pdu: %s", err)
	} else {
		return packet, pdu, nil
	}
}

// Goroutine dedicated to handling all incoming response packets
func (self *Client) recvLoop() {
	for {
		//self.log.Printf("recv...\n")

		if _, pdu, err := self.recv(); err != nil {
			self.log.Printf("recv: %s\n", err)
		} else {
			//self.log.Printf("recv: %+v\n", pdu)

			self.dispatchResponse(pdu)
		}
	}
}

// Dispatch response PDU to waiting dispatchRequest()
func (self *Client) dispatchResponse(pdu PDU) {
	self.requestMutex.Lock()
	defer self.requestMutex.Unlock()

	if request, ok := self.requests[pdu.RequestID]; !ok {
		self.log.Printf("recv: request %d: unknown\n", pdu.RequestID)
	} else {
		select {
		case request.responseChan <- pdu:
			//self.log.Printf("recv: request %v: response\n", request)
		default:
			self.log.Printf("recv: request %v: blocked\n", request)
		}

	}
}

// Allocate RequestID and start tracking responses
func (self *Client) startRequest(request *request) {
	self.requestMutex.Lock()
	defer self.requestMutex.Unlock()

	// allocate request ID
	request.pdu.RequestID = self.requestId

	self.requests[request.pdu.RequestID] = request

	self.requestId++

	//self.log.Printf("request %v: start\n", request)
}

// Stop tracking responses for given request
func (self *Client) finishRequest(request *request) {
	self.requestMutex.Lock()
	defer self.requestMutex.Unlock()

	delete(self.requests, request.pdu.RequestID)
}

// Start request, retrying send until response or timeout.
// Uses startRequest/finishRequest to register for dispatchResponse()
func (self *Client) request(request *request) (PDU, error) {
	self.startRequest(request)

	defer self.finishRequest(request)

	for retry := self.retries; retry > 0; retry-- {
		//self.log.Printf("request %v: send\n", request)

		if err := self.sendPDU(request.packet, request.pdu); err != nil {
			return PDU{}, err
		}

		timeout := time.After(self.timeout)

		select {
		case <-timeout:
			//self.log.Printf("request %v: retry\n", request)
			continue
		case responsePDU := <-request.responseChan:
			//self.log.Printf("request %v: response\n", request)
			return responsePDU, nil
		}
	}

	return PDU{}, fmt.Errorf("timeout")
}

// Build a generic-PDU request, dispatch it, and return any errors
func (self *Client) requestPDU(requestType wapsnmp.BERType, requestPDU PDU) (PDU, error) {
	request := request{
		packet: Packet{
			Version:   self.version,
			Community: self.community,
			PduType:   requestType,
		},
		pdu:          requestPDU,
		responseChan: make(chan PDU),
	}

	// send and wait for response
	if responsePDU, err := self.request(&request); err != nil {
		//self.log.Printf("request %v: error: %s\n", request, err)

		return responsePDU, err

	} else if responsePDU.ErrorStatus != 0 {
		err := RequestError{requestType, responsePDU.ErrorStatus, responsePDU.ErrorIndex}

		//self.log.Printf("request %v: SNMP error: %s\n", request, err)

		return responsePDU, err

	} else {
		//self.log.Printf("request %v: done\n", request)

		return responsePDU, nil
	}
}

// Dispatch a generic-PDU request, with the given OIDs as VarBinds
func (self *Client) requestGet(requestType wapsnmp.BERType, oids []OID) ([]VarBind, error) {
	requestPDU := PDU{}

	for _, oid := range oids {
		requestPDU.VarBinds = append(requestPDU.VarBinds, VarBind{Name: wapsnmp.Oid(oid)})
	}

	if responsePDU, err := self.requestPDU(requestType, requestPDU); err != nil {
		return nil, err
	} else if len(responsePDU.VarBinds) != len(requestPDU.VarBinds) {
		return nil, fmt.Errorf("response var-binds mismatch")
	} else {
		// TODO: noSuchObject, noSuchInstance
		return responsePDU.VarBinds, nil
	}
}

func (self *Client) Get(oids ...OID) ([]VarBind, error) {
	self.log.Printf("Get %v\n", oids)

	return self.requestGet(wapsnmp.AsnGetRequest, oids)
}

func (self *Client) GetNext(oids ...OID) ([]VarBind, error) {
	self.log.Printf("GetNext %v\n", oids)

	return self.requestGet(wapsnmp.AsnGetNextRequest, oids)
}

// Get a scalar SNMP object, returning its value
// Returns nil if the object is not found
func (self *Client) GetObject(object *Object) (interface{}, error) {
	if varBinds, err := self.Get(object.OID.define(0)); err != nil {
		return nil, err
	} else {
		for _, varBind := range varBinds {
			oid := OID(varBind.Name)

			if index := object.Index(oid); index == nil {
				return nil, fmt.Errorf("response var-bind OID mismatch")
			} else if varBind.Value == wapsnmp.NoSuchObject || varBind.Value == wapsnmp.NoSuchInstance {
				return nil, nil
			} else if objectValue, err := object.ParseValue(varBind.Value); err != nil {
				return nil, err
			} else {
				return objectValue, nil
			}
		}

		return nil, nil
	}
}

// Probe for supported MIBS
func (self *Client) ProbeMIBs(handler func(*MIB)) error {
	for _, mib := range mibs {
		// TODO: probe...
		handler(mib)
	}

	return nil
}

// Probe for supported Objects in given MIB
func (self *Client) ProbeMIBObjects(mib *MIB, handler func(*Object)) error {
	for _, object := range mib.objects {
		if object.Table != nil {
			continue
		}

		// TODO: probe...
		handler(object)
	}

	return nil
}

// Probe for supported Tables in given MIB
func (self *Client) ProbeMIBTables(mib *MIB, handler func(*Table)) error {
	for _, table := range mib.tables {
		// TODO: probe...
		handler(table)
	}

	return nil
}
