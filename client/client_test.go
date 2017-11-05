package client

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type testLogger struct {
	t      *testing.T
	prefix string
}

func (logger testLogger) Printf(format string, args ...interface{}) {
	logger.t.Logf(logger.prefix+format, args...)
}

func makeTestClient(t *testing.T) (*testTransport, *Client) {
	var testTransport = testTransport{
		recvChan: make(chan IO),
	}
	var client = makeClient(Logging{
		Debug: testLogger{t, "DEBUG: "},
		Info:  testLogger{t, "INFO: "},
		Warn:  testLogger{t, "WARN: "},
		Error: testLogger{t, "Error: "},
	})

	client.version = SNMPVersion
	client.community = []byte("public")
	client.transport = &testTransport

	return &testTransport, &client
}

type testTransport struct {
	mock.Mock

	recvChan      chan IO
	recvErrorChan chan error
}

func (transport *testTransport) Send(io IO) error {
	args := transport.MethodCalled(io.PDUType.String(), io)

	transport.recvChan <- args.Get(1).(IO)

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

func (transport *testTransport) mockGetNext(requestID requestID, oid snmp.OID, varBind snmp.VarBind) {
	transport.On("GetNextRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetNextRequestType,
		PDU: snmp.PDU{
			RequestID: int(requestID),
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.PDU{
			RequestID: int(requestID),
			VarBinds: []snmp.VarBind{
				varBind,
			},
		},
	})
}

func assertVarBind(t *testing.T, varBinds []snmp.VarBind, index int, expectedOID snmp.OID, expectedValue interface{}) {
	if len(varBinds) < index {
		t.Errorf("VarBinds[%d]: short %d", index, len(varBinds))
	} else if value, err := varBinds[index].Value(); err != nil {
		t.Errorf("VarBinds[%d]: invalid Value: %v", index, err)
	} else {
		assert.Equal(t, expectedOID, varBinds[index].OID())
		assert.Equal(t, expectedValue, value)
	}
}

func TestGetRequest(t *testing.T) {
	var testTransport, client = makeTestClient(t)
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var value = []byte("qmsk-snmp test")

	go client.Run()
	defer client.Close()

	testTransport.On("GetRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.PDU{
			RequestID: 1,
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.PDU{
			RequestID: 1,
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, value),
			},
		},
	})

	if varBinds, err := client.Get(oid); err != nil {
		t.Fatalf("Get(%v): %v", oid, err)
	} else if !testTransport.AssertExpectations(t) {

	} else {
		assertVarBind(t, varBinds, 0, oid, value)
	}
}

func TestWalk(t *testing.T) {
	var testTransport, client = makeTestClient(t)
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1} // IF-MIB::ifName
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if1")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if2")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 2, 1}, snmp.Counter32(0)),
	}

	testTransport.mockGetNext(1, oid, varBinds[0])
	testTransport.mockGetNext(2, varBinds[0].OID(), varBinds[1])
	testTransport.mockGetNext(3, varBinds[1].OID(), varBinds[2])

	go client.Run()
	defer client.Close()

	if err := client.Walk(func(varBinds ...snmp.VarBind) error {

		return nil
	}, oid); err != nil {
		t.Fatalf("Walk(%v): %v", oid, err)
	}
}
