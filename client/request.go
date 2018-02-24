package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

type requestID uint32
type requestKey struct {
	id   requestID
	addr string
}

type requestMap map[requestKey]*request

type request struct {
	send      IO
	id        requestID
	timeout   time.Duration
	retry     uint
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

func (request *request) Error() error {
	if request.recv.PDU.ErrorStatus != 0 {
		return SNMPError{
			RequestType:  request.send.PDUType,
			ResponseType: request.recv.PDUType,
			ResponsePDU:  request.recv.PDU,
		}
	} else {
		return nil
	}
}

func (request *request) wait() error {
	if err, ok := <-request.waitChan; !ok {
		return fmt.Errorf("request canceled")
	} else {
		return err
	}
}

func (request *request) start(timeout time.Duration, timeoutChan chan requestID) {
	request.timer = time.AfterFunc(timeout, func() {
		timeoutChan <- request.id
	})
}

func (request *request) close() {
	if request.timer != nil {
		request.timer.Stop()
	}
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
	ResponsePDU  snmp.PDU
}

func (err SNMPError) Error() string {
	return fmt.Sprintf("SNMP %v error: %v @ %v", err.RequestType, err.ResponsePDU.ErrorStatus, err.ResponsePDU.ErrorVarBind())
}
