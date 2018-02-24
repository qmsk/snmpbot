package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/qmsk/snmpbot/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
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
	var client = makeClient(logging.Logging{
		Debug: testLogger{t, "DEBUG: "},
		Info:  testLogger{t, "INFO: "},
		Warn:  testLogger{t, "WARN: "},
		Error: testLogger{t, "Error: "},
	})

	client.version = SNMPVersion
	client.community = []byte("public")
	client.timeout = 10 * time.Millisecond
	client.transport = &testTransport
	client.maxVars = DefaultMaxVars

	return &testTransport, &client
}

func withTestClient(t *testing.T, f func(*testTransport, *Client)) {
	var transport, client = makeTestClient(t)

	go client.Run()
	defer client.Close()

	f(transport, client)

	transport.AssertExpectations(t)
}

type testTransport struct {
	mock.Mock

	recvChan      chan IO
	recvErrorChan chan error
}

func (transport *testTransport) String() string {
	return "<test>"
}

func (transport *testTransport) Send(io IO) error {
	var requestID = io.PDU.RequestID

	io.PDU.RequestID = 0

	args := transport.MethodCalled(io.PDUType.String(), io)

	if ret := args.Get(1); ret == nil {
		// no response
	} else {
		recv := ret.(IO)
		recv.PDU.RequestID = requestID

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

func (transport *testTransport) mockGetTimeout(oid snmp.OID) {
	transport.On("GetRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.PDU{
			VarBinds: []snmp.VarBind{
				snmp.MakeVarBind(oid, nil),
			},
		},
	}).Return(error(nil), nil)
}

func (transport *testTransport) mockGet(oid snmp.OID, varBind snmp.VarBind) {
	transport.On("GetRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.PDU{
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
			VarBinds: []snmp.VarBind{
				varBind,
			},
		},
	})
}

func (transport *testTransport) mockGetMany(oids []snmp.OID, varBinds []snmp.VarBind) {
	var reqVars = make([]snmp.VarBind, len(oids))
	for i, oid := range oids {
		reqVars[i] = snmp.MakeVarBind(oid, nil)
	}

	transport.On("GetRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetRequestType,
		PDU: snmp.PDU{
			VarBinds: reqVars,
		},
	}).Return(error(nil), IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.PDU{
			VarBinds: varBinds,
		},
	})
}

func (transport *testTransport) mockGetNext(oid snmp.OID, varBind snmp.VarBind) {
	transport.On("GetNextRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetNextRequestType,
		PDU: snmp.PDU{
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
			VarBinds: []snmp.VarBind{
				varBind,
			},
		},
	})
}

func (transport *testTransport) mockGetNextMulti(oids []snmp.OID, varBinds []snmp.VarBind) {
	var requestVars = make([]snmp.VarBind, len(oids))
	for i, oid := range oids {
		requestVars[i] = snmp.MakeVarBind(oid, nil)
	}

	transport.On("GetNextRequest", IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetNextRequestType,
		PDU: snmp.PDU{
			VarBinds: requestVars,
		},
	}).Return(error(nil), IO{
		Packet: snmp.Packet{
			Version:   snmp.SNMPv2c,
			Community: []byte("public"),
		},
		PDUType: snmp.GetResponseType,
		PDU: snmp.PDU{
			VarBinds: varBinds,
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
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var value = []byte("qmsk-snmp test")

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.mockGet(oid, snmp.MakeVarBind(oid, value))

		if varBinds, err := client.Get(oid); err != nil {
			t.Fatalf("Get(%v): %v", oid, err)
		} else {
			assertVarBind(t, varBinds, 0, oid, value)
		}
	})
}

func TestGetTimeout(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.mockGetTimeout(oid)

		if varBinds, err := client.Get(oid); err == nil {
			t.Errorf("Get(%v): %v", oid, varBinds)
		} else if timeoutErr, ok := err.(TimeoutError); !ok {
			t.Errorf("Get(%v): %v", oid, err)
		} else {
			assert.EqualError(t, err, fmt.Sprintf("SNMP <test> timeout for GetRequest<1.3.6.1.2.1.1.5.0>@1 after %v", timeoutErr.Duration))
		}
	})
}

func TestGetSendError(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var err = fmt.Errorf("Send error")

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.On("GetRequest", IO{
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetRequestType,
			PDU: snmp.PDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(oid, nil),
				},
			},
		}).Return(err, nil)

		if varBinds, err := client.Get(oid); err == nil {
			t.Errorf("Get(%v): %v", oid, varBinds)
		} else {
			assert.EqualError(t, err, "SNMP <test> send failed: Send error")
		}
	})
}

func TestGetRequestErrorValue(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var value = snmp.NoSuchObjectValue

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.mockGet(oid, snmp.MakeVarBind(oid, value))

		if varBinds, err := client.Get(oid); err != nil {
			t.Fatalf("Get(%v): %v", oid, err)
		} else {
			assertVarBind(t, varBinds, 0, oid, value)
		}
	})
}

func TestGetRequestBig(t *testing.T) {
	var oids = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 3},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 4},
	}
	var values = [][]byte{
		[]byte("qmsk-snmp test 0"),
		[]byte("qmsk-snmp test 1"),
		[]byte("qmsk-snmp test 2"),
		[]byte("qmsk-snmp test 3"),
		[]byte("qmsk-snmp test 4"),
	}

	withTestClient(t, func(transport *testTransport, client *Client) {
		client.maxVars = 2

		transport.mockGetMany([]snmp.OID{oids[0], oids[1]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[0], values[0]),
			snmp.MakeVarBind(oids[1], values[1]),
		})
		transport.mockGetMany([]snmp.OID{oids[2], oids[3]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[2], values[2]),
			snmp.MakeVarBind(oids[3], values[3]),
		})
		transport.mockGetMany([]snmp.OID{oids[4]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[4], values[4]),
		})

		if varBinds, err := client.Get(oids...); err != nil {
			t.Fatalf("Get(%v): %v", oids, err)
		} else {
			for i, oid := range oids {
				assertVarBind(t, varBinds, i, oid, values[i])
			}
		}
	})
}

func TestGetNothing(t *testing.T) {
	withTestClient(t, func(transport *testTransport, client *Client) {
		if varBinds, err := client.Get(); err != nil {
			t.Fatalf("Get(): %v", err)
		} else {
			assert.Equal(t, []snmp.VarBind(nil), varBinds)
		}
	})
}

func TestWalk(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1} // IF-MIB::ifName
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if1")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if2")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 2, 1}, snmp.Counter32(0)),
	}

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.mockGetNext(oid, varBinds[0])
		transport.mockGetNext(varBinds[0].OID(), varBinds[1])
		transport.mockGetNext(varBinds[1].OID(), varBinds[2])

		if err := client.Walk(func(varBinds ...snmp.VarBind) error {

			return nil
		}, oid); err != nil {
			t.Fatalf("Walk(%v): %v", oid, err)
		}
	})
}

func TestWalkV2(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1} // IF-MIB::ifName
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if1")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if2")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, snmp.EndOfMibViewValue),
	}

	withTestClient(t, func(transport *testTransport, client *Client) {
		transport.mockGetNext(oid, varBinds[0])
		transport.mockGetNext(varBinds[0].OID(), varBinds[1])
		transport.mockGetNext(varBinds[1].OID(), varBinds[2])

		if err := client.Walk(func(varBinds ...snmp.VarBind) error {

			return nil
		}, oid); err != nil {
			t.Fatalf("Walk(%v): %v", oid, err)
		}
	})
}

func TestWalkPartial(t *testing.T) {
	var oid1 = snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 1} // Q-BRIDGE-MIB::dot1qTpFdbAddress (not-accessible)
	var oid2 = snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2} // Q-BRIDGE-MIB::dot1qTpFdbPort

	var errBind = snmp.MakeVarBind(oid1, snmp.EndOfMibViewValue)
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2, 1, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}, int(1)),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2, 1, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}, int(3)),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 3, 1, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}, int(1)),
	}

	withTestClient(t, func(transport *testTransport, client *Client) {
		var walkMock mock.Mock
		defer walkMock.AssertExpectations(t)

		transport.mockGetNextMulti([]snmp.OID{oid1, oid2}, []snmp.VarBind{varBinds[0], varBinds[0]})
		walkMock.On("walk[1/2]", errBind)
		walkMock.On("walk[2/2]", varBinds[0])
		transport.mockGetNextMulti([]snmp.OID{oid1, varBinds[0].OID()}, []snmp.VarBind{varBinds[0], varBinds[1]})
		walkMock.On("walk[1/2]", errBind)
		walkMock.On("walk[2/2]", varBinds[1])
		transport.mockGetNextMulti([]snmp.OID{oid1, varBinds[1].OID()}, []snmp.VarBind{varBinds[0], varBinds[2]})

		if err := client.Walk(func(varBinds ...snmp.VarBind) error {
			for i, varBind := range varBinds {
				walkMock.MethodCalled(fmt.Sprintf("walk[%d/%d]", i+1, len(varBinds)), varBind)
			}

			return nil
		}, oid1, oid2); err != nil {
			t.Fatalf("Walk(%v, %v): %v", oid1, oid2, err)
		}
	})
}
