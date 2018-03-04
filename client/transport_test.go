package client

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/mock"
	"net"
)

type testAddr string

func (addr testAddr) Network() string {
	return "test"
}

func (addr testAddr) String() string {
	return string(addr)
}

func makeTestTransport() testTransport {
	return testTransport{
		recvChan: make(chan IO),
	}
}

type testTransport struct {
	mock.Mock

	recvChan      chan IO
	recvErrorChan chan error
}

func (transport *testTransport) String() string {
	return "testing"
}

func (transport *testTransport) Resolve(addr string) (net.Addr, error) {
	return testAddr(addr), nil
}

func (transport *testTransport) Send(io IO) error {
	var requestID = io.PDU.GetRequestID()

	// override requestID to 0 for assert.Equal() comparison
	io.PDU.SetRequestID(0)

	args := transport.MethodCalled(io.PDUType.String(), io)

	if ret := args.Get(1); ret == nil {
		// no response
	} else {
		recv := ret.(IO)
		recv.PDU.SetRequestID(requestID)

		transport.recvChan <- recv
	}

	return args.Error(0)
}

func (transport *testTransport) Recv() (IO, error) {
	select {
	case io, ok := <-transport.recvChan:
		if ok {
			return io, nil
		} else {
			return io, EOF
		}
	case err := <-transport.recvErrorChan:
		return IO{}, err
	}
}

func (transport *testTransport) Close() error {
	close(transport.recvChan)

	return nil
}

func (transport *testTransport) mockGetTimeout(addr string, oid snmp.OID) {
	transport.On("GetRequest", IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.GenericPDU{
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), nil)
}

func (transport *testTransport) mockGet(addr string, oid snmp.OID, varBind snmp.VarBind) {
	transport.On("GetRequest", IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.GenericPDU{
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.GenericPDU{
			VarBinds: []snmp.VarBind{
				varBind,
			},
		},
	})
}

func (transport *testTransport) mockGetMany(addr string, oids []snmp.OID, varBinds []snmp.VarBind) {
	var reqVars = make([]snmp.VarBind, len(oids))
	for i, oid := range oids {
		reqVars[i] = snmp.MakeVarBind(oid, nil)
	}

	transport.On("GetRequest", IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.GenericPDU{
			VarBinds: reqVars,
		},
	}).Return(error(nil), IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.GenericPDU{
			VarBinds: varBinds,
		},
	})
}

func (transport *testTransport) mockGetNext(addr string, oid snmp.OID, varBind snmp.VarBind) {
	transport.On("GetNextRequest", IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetNextRequestType,
		PDU: snmp.GenericPDU{
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.GenericPDU{
			VarBinds: []snmp.VarBind{
				varBind,
			},
		},
	})
}

func (transport *testTransport) mockGetNextMulti(addr string, oids []snmp.OID, varBinds []snmp.VarBind) {
	var requestVars = make([]snmp.VarBind, len(oids))
	for i, oid := range oids {
		requestVars[i] = snmp.MakeVarBind(oid, nil)
	}

	transport.On("GetNextRequest", IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetNextRequestType,
		PDU: snmp.GenericPDU{
			VarBinds: requestVars,
		},
	}).Return(error(nil), IO{
		Addr: testAddr(addr),
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.GenericPDU{
			VarBinds: varBinds,
		},
	})
}
