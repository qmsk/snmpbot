package client

import (
	"fmt"
	snmp "github.com/qmsk/snmpbot/snmp_new"
)

type requestID uint32

type request struct {
	send     IO
	waitChan chan error
	recv     IO
}

func (request request) String() string {
	return fmt.Sprintf("%s<%s>: %s", request.send.PDUType.String(), request.send.PDU.String(), request.recv.PDU.String())
}

func (request *request) wait() (IO, error) {
	if err, ok := <-request.waitChan; !ok {
		return request.recv, fmt.Errorf("request canceled")
	} else {
		return request.recv, err
	}
}

func (request *request) cancel() {
	close(request.waitChan)
}

func (request *request) fail(err error) {
	request.waitChan <- err
}

func (request *request) done(recv IO) {
	request.recv = recv
	request.waitChan <- nil
}

type SNMPError struct {
	RequestType  snmp.PDUType
	ResponseType snmp.PDUType
	ErrorStatus  snmp.ErrorStatus
	VarBind      snmp.VarBind
}

func (err SNMPError) Error() string {
	return fmt.Sprintf("SNMP %v error: %v @ %v", err.RequestType, err.ErrorStatus, err.VarBind)
}

func (client *Client) request(send IO) (IO, error) {
	var request = request{
		send:     send,
		waitChan: make(chan error, 1),
	}

	client.requestChan <- &request

	if recv, err := request.wait(); err != nil {
		return recv, err
	} else if recv.PDU.ErrorStatus != 0 {
		return recv, SNMPError{
			RequestType:  send.PDUType,
			ResponseType: recv.PDUType,
			ErrorStatus:  recv.PDU.ErrorStatus,
			VarBind:      recv.PDU.VarBinds[recv.PDU.ErrorIndex], // XXX
		}
	} else {
		client.log.Infof("Request %v", request)

		return recv, nil
	}
}

func (client *Client) requestRead(requestType snmp.PDUType, varBinds []snmp.VarBind) (snmp.PDUType, []snmp.VarBind, error) {
	var send = IO{
		Addr: client.addr,
		Packet: snmp.Packet{
			Version:   client.version,
			Community: client.community,
		},
		PDUType: requestType,
		PDU: snmp.PDU{
			VarBinds: varBinds,
		},
	}

	if recv, err := client.request(send); err != nil {
		return recv.PDUType, recv.PDU.VarBinds, err
	} else {
		return recv.PDUType, recv.PDU.VarBinds, nil
	}
}

func (client *Client) Get(OIDs ...snmp.OID) ([]snmp.VarBind, error) {
	var requestVars = make([]snmp.VarBind, len(OIDs))

	for i, oid := range OIDs {
		requestVars[i].Name = oid
		requestVars[i].SetNull()
	}

	if responseType, responseVars, err := client.requestRead(snmp.GetRequestType, requestVars); err != nil {
		return responseVars, err
	} else if responseType != snmp.GetResponseType {
		return responseVars, fmt.Errorf("Unexpected response type %v for GetRequest", responseType)
	} else {
		return responseVars, nil
	}
}

func (client *Client) GetNext(OIDs ...snmp.OID) ([]snmp.VarBind, error) {
	var requestVars = make([]snmp.VarBind, len(OIDs))

	for i, oid := range OIDs {
		requestVars[i].Name = oid
		requestVars[i].SetNull()
	}

	if responseType, responseVars, err := client.requestRead(snmp.GetNextRequestType, requestVars); err != nil {
		return responseVars, err
	} else if responseType != snmp.GetResponseType {
		return responseVars, fmt.Errorf("Unexpected response type %v for GetNextRequest", responseType)
	} else {
		return responseVars, nil
	}
}
