package client

import (
	"fmt"
	snmp "github.com/qmsk/snmpbot/snmp_new"
	"net"
)

func makeTestServer() *testServer {
	var testServer = testServer{
		values: make(map[string]interface{}),
	}

	if udp, err := ListenUDP("127.0.0.1:0", UDPOptions{}); err != nil {
		panic(err)
	} else {
		testServer.udp = udp
	}

	if udpAddr, err := testServer.udp.LocalAddr(); err != nil {
		panic(err)
	} else {
		testServer.udpAddr = udpAddr
	}

	return &testServer
}

type testServer struct {
	udp     *UDP
	udpAddr *net.UDPAddr

	values map[string]interface{}
}

func (testServer *testServer) MockGet(oid snmp.OID, value interface{}) {
	testServer.values[oid.String()] = value
}

func (testServer *testServer) get(oid snmp.OID) (snmp.VarBind, error) {
	var varBind = snmp.MakeVarBind(oid, nil)

	if value, ok := testServer.values[oid.String()]; ok {
		varBind.Set(value)
	} else {
		varBind.SetError(snmp.NoSuchObjectValue)
	}

	return varBind, nil
}

func (testServer *testServer) handleGet(pdu snmp.PDU) (snmp.PDU, error) {
	var response = snmp.PDU{
		RequestID: pdu.RequestID,
		VarBinds:  make([]snmp.VarBind, len(pdu.VarBinds)),
	}

	for i, get := range pdu.VarBinds {
		if varBind, err := testServer.get(get.OID()); err == nil {
			response.VarBinds[i] = varBind
		} else if errorStatus, ok := err.(snmp.ErrorStatus); ok {
			response.ErrorStatus = errorStatus
			response.ErrorIndex = i
			response.VarBinds[i] = get
		} else {
			return response, err
		}
	}

	return response, nil
}

func (testServer *testServer) handle(recv IO) (send IO, err error) {
	send.Addr = recv.Addr
	send.Packet.Version = recv.Packet.Version
	send.Packet.Community = recv.Packet.Community

	switch recv.PDUType {
	case snmp.GetRequestType:
		send.PDUType = snmp.GetResponseType
		send.PDU, err = testServer.handleGet(recv.PDU)
	default:
		return send, fmt.Errorf("Invalid request PDU type: %v", recv.PDUType)
	}

	return send, nil
}

func (testServer *testServer) run() {
	for {
		if recv, err := testServer.udp.Recv(); err != nil {
			panic(err)
		} else if send, err := testServer.handle(recv); err != nil {
			panic(err)
		} else if err := testServer.udp.Send(send); err != nil {
			panic(err)
		}
	}
}

func (testServer *testServer) stop() {
	testServer.udp.conn.Close()
}
