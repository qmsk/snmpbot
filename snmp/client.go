package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "log"
    "net"
    "os"
    "sync"
    "time"
    wapsnmp "github.com/cdevr/WapSNMP"
)

const (
    TIMEOUT     = time.Duration(2) * time.Second
    RETRIES     = 3
    VERSION     = wapsnmp.SNMPv2c
)

type Client struct {
    log     *log.Logger

    version     wapsnmp.SNMPVersion
    community   string
    timeout     time.Duration
    retries     int

    udpMutex    sync.Mutex
    udpConn    *net.UDPConn
    udpSize     uint

    requestMutex    sync.Mutex
    requestId       int32
    requests        map[int32]*request

    gosnmp  *gosnmp.GoSNMP
}

func (self Client) String() string {
    return fmt.Sprintf("%s", self.udpConn.RemoteAddr())
}

type request struct {
    packet          Packet
    pdu             PDU

    responseChan    chan PDU
}

func (self request) String() string {
    return fmt.Sprintf("%d", self.pdu.RequestID)
}

type RequestError struct {
    RequestType     wapsnmp.BERType
    ErrorStatus     int
    ErrorIndex      int
}

func (self RequestError) Error() string {
    return fmt.Sprintf("SNMP error:%d [%d]", self.ErrorStatus, self.ErrorIndex)
}

func Connect(config Config) (*Client, error) {
    client := &Client{
        gosnmp:   &gosnmp.GoSNMP{
            Target:     config.Host,
            Port:       161,
            Version:    gosnmp.Version2c,
            Community:  config.Community,
            Timeout:    TIMEOUT,
            Retries:    RETRIES,
        },

        version:    VERSION,
        community:  config.Community,
        timeout:    TIMEOUT,
        retries:    RETRIES,

        udpSize:    UDP_SIZE,

        requestId:  1337,
        requests:   make(map[int32]*request),
    }

    if err := client.gosnmp.Connect(); err != nil {
        return nil, err
    }

    if err := client.connect(config); err != nil {
        return nil, err
    }

    // process responses
    go client.recvLoop()

    return client, nil
}

func (self *Client) connect(config Config) error {
    netAddr := net.JoinHostPort(config.Host, config.Port)

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

func (self *Client) send(packet Packet, pdu PDU) error {
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

// Go recv() and dispatch responses
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

func (self *Client) request(requestType wapsnmp.BERType, varBinds []VarBind) ([]VarBind, error) {
    request := request{
        packet: Packet{
            Version:    self.version,
            Community:  self.community,
            PduType:    requestType,
        },
        pdu: PDU{
            VarBinds:   varBinds,
        },
        responseChan: make(chan PDU),
    }

    // send and wait for response
    var responsePDU PDU
    var retry int

    self.startRequest(&request)
    defer self.finishRequest(&request)

retry:
    for retry = self.retries; retry > 0; retry-- {
        //self.log.Printf("request %v: send\n", request)

        if err := self.send(request.packet, request.pdu); err != nil {
            return nil, err
        }

        timeout := time.After(self.timeout)

        select {
        case <-timeout:
            //self.log.Printf("request %v: retry\n", request)
            continue
        case responsePDU = <-request.responseChan:
            //self.log.Printf("request %v: response\n", request)
            // leaves retry > 0
            break retry
        }
    }

    // handle response
    if retry == 0 {
        //self.log.Printf("request %v: timeout\n", request)

        return nil, fmt.Errorf("timeout")

    } else if responsePDU.ErrorStatus != 0 {
        err := RequestError{requestType, responsePDU.ErrorStatus, responsePDU.ErrorIndex}

        //self.log.Printf("request %v: error: %s\n", request, err)

        return nil, err
    } else {
        //self.log.Printf("request %v: done\n", request)

        return responsePDU.VarBinds, nil
    }
}

func (self *Client) Get(oids... OID) ([]VarBind, error) {
    var requestVars []VarBind

    self.log.Printf("Get %v\n", oids)

    for _, oid := range oids {
        requestVars = append(requestVars, VarBind{Name: wapsnmp.Oid(oid)})
    }

    if responseVars, err := self.request(wapsnmp.AsnGetRequest, requestVars); err != nil {
        return nil, err
    } else if len(responseVars) != len(requestVars) {
        return nil, fmt.Errorf("response var-binds count mismatch")
    } else {
        return responseVars, nil
    }
}

func (self *Client) GetNext(oids... OID) ([]VarBind, error) {
    var requestVars []VarBind

    self.log.Printf("GetNext %v\n", oids)

    for _, oid := range oids {
        requestVars = append(requestVars, VarBind{Name: wapsnmp.Oid(oid)})
    }

    if responseVars, err := self.request(wapsnmp.AsnGetNextRequest, requestVars); err != nil {
        return nil, err
    } else if len(responseVars) != len(requestVars) {
        return nil, fmt.Errorf("response var-binds count mismatch")
    } else {
        return responseVars, nil
    }
}

func (self *Client) Walk(walkOID OID, handler func (oid OID, value interface{})) error {
    nextOID := walkOID.Copy()

    for {
        if varBinds, err := self.GetNext(nextOID); err != nil {
            return err
        } else {
            varBind := varBinds[0]
            varOID := OID(varBind.Name)

            self.log.Printf("Walk %v: %v\n", walkOID, varOID)

            if varBind.Value == wapsnmp.EndOfMibView {
                break
            } else if varOID.Equals(nextOID) {
                break
            } else if walkOID.Index(varOID) == nil {
                break
            } else {
                nextOID = varOID

                handler(varOID, varBind.Value)
            }
        }
    }

    return nil
}
