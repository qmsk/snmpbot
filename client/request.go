package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

type requestID uint32

type request struct {
	send      IO
	id        requestID
	timeout   time.Duration
	retry     int
	startTime time.Time
	timer     *time.Timer
	waitChan  chan error
	recv      IO
	recvOK    bool
}

func (request request) String() string {
	if request.recvOK {
		return fmt.Sprintf("%s<%s>@%d: %s", request.send.PDUType.String(), request.send.PDU.String(), request.id, request.recv.PDU.String())
	} else {
		return fmt.Sprintf("%s<%s>@%d", request.send.PDUType.String(), request.send.PDU.String(), request.id)
	}
}

func (request *request) wait() (IO, error) {
	if err, ok := <-request.waitChan; !ok {
		return request.recv, fmt.Errorf("request canceled")
	} else {
		return request.recv, err
	}
}

func (request *request) start(timeout time.Duration, timeoutChan chan requestID) {
	request.timer = time.AfterFunc(timeout, func() {
		timeoutChan <- request.id
	})
}

func (request *request) close() {
	request.timer.Stop()
	close(request.waitChan)
}

func (request *request) fail(err error) {
	request.waitChan <- err
	request.close()
}

func (request *request) done(recv IO) {
	request.recv = recv
	request.recvOK = true
	request.waitChan <- nil
	request.close()
}

func (request *request) failTimeout(transport Transport) {
	request.fail(TimeoutError{
		transport: transport,
		request:   request,
		Duration:  time.Now().Sub(request.startTime),
	})
}

type TimeoutError struct {
	transport Transport
	request   *request
	Duration  time.Duration
}

func (err TimeoutError) Error() string {
	return fmt.Sprintf("SNMP %v timeout for %v after %v", err.transport, err.request, err.Duration)
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
		send:      send,
		timeout:   client.timeout,
		retry:     client.retry,
		startTime: time.Now(),
		waitChan:  make(chan error, 1),
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
