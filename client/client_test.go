package client

import (
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func makeTestClient(t *testing.T, addr string) (*testTransport, *Engine, *Client) {
	SetLogging(logging.TestLogging(t))

	var options = Options{
		Community: "public",
		Timeout:   10 * time.Millisecond,
	}

	var testTransport = makeTestTransport()
	var engine = makeEngine(&testTransport)
	var client = makeClient(&engine, options)

	engine.log = logging.WithPrefix(log, fmt.Sprintf("Engine<%v>", &engine))

	client.addr = testAddr(addr)
	client.log = logging.WithPrefix(log, fmt.Sprintf("Client<%v>", &client))

	return &testTransport, &engine, &client
}

func withTestClient(t *testing.T, addr string, f func(*testTransport, *Client)) {
	var transport, engine, client = makeTestClient(t, addr)

	go engine.Run()
	defer engine.Close()

	f(transport, client)

	transport.AssertExpectations(t)
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

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGet("test", oid, snmp.MakeVarBind(oid, value))

		if varBinds, err := client.Get(oid); err != nil {
			t.Fatalf("Get(%v): %v", oid, err)
		} else {
			assertVarBind(t, varBinds, 0, oid, value)
		}
	})
}

func TestGetTimeout(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetTimeout("test", oid)

		if varBinds, err := client.Get(oid); err == nil {
			t.Errorf("Get(%v): %v", oid, varBinds)
		} else if timeoutErr, ok := err.(TimeoutError); !ok {
			t.Errorf("Get(%v): %v", oid, err)
		} else {
			assert.EqualError(t, err, fmt.Sprintf("SNMP<testing> timeout for GetRequest<1.3.6.1.2.1.1.5.0>@test[1] after %v", timeoutErr.Duration))
		}
	})
}

func TestGetSendError(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var err = fmt.Errorf("Send error")

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.On("GetRequest", IO{
			Addr: testAddr("test"),
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
			assert.EqualError(t, err, "SNMP<testing> send failed: Send error")
		}
	})
}

func TestGetRecvWrongAddr(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.On("GetRequest", IO{
			Addr: testAddr("test"),
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
		}).Return(nil, IO{
			Addr: testAddr("test2"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.PDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(oid, 1),
				},
			},
		})

		if varBinds, err := client.Get(oid); err == nil {
			t.Errorf("Get(%v): %v", oid, varBinds)
		} else if timeoutErr, ok := err.(TimeoutError); !ok {
			t.Errorf("Get(%v): %v", oid, err)
		} else {
			assert.EqualError(t, err, fmt.Sprintf("SNMP<testing> timeout for GetRequest<1.3.6.1.2.1.1.5.0>@test[1] after %v", timeoutErr.Duration))
		}
	})
}

func TestGetRecvWrongCommunity(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.On("GetRequest", IO{
			Addr: testAddr("test"),
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
		}).Return(nil, IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("not-public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.PDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(oid, 1),
				},
			},
		})

		if varBinds, err := client.Get(oid); err == nil {
			t.Errorf("Get(%v): %v", oid, varBinds)
		} else if timeoutErr, ok := err.(TimeoutError); !ok {
			t.Errorf("Get(%v): %v", oid, err)
		} else {
			assert.EqualError(t, err, fmt.Sprintf("SNMP<testing> timeout for GetRequest<1.3.6.1.2.1.1.5.0>@test[1] after %v", timeoutErr.Duration))
		}
	})
}

func TestGetRequestErrorValue(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0}
	var value = snmp.NoSuchObjectValue

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGet("test", oid, snmp.MakeVarBind(oid, value))

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

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 2

		transport.mockGetMany("test", []snmp.OID{oids[0], oids[1]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[0], values[0]),
			snmp.MakeVarBind(oids[1], values[1]),
		})
		transport.mockGetMany("test", []snmp.OID{oids[2], oids[3]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[2], values[2]),
			snmp.MakeVarBind(oids[3], values[3]),
		})
		transport.mockGetMany("test", []snmp.OID{oids[4]}, []snmp.VarBind{
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
	withTestClient(t, "test", func(transport *testTransport, client *Client) {
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

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNext("test", oid, varBinds[0])
		transport.mockGetNext("test", varBinds[0].OID(), varBinds[1])
		transport.mockGetNext("test", varBinds[1].OID(), varBinds[2])

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

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNext("test", oid, varBinds[0])
		transport.mockGetNext("test", varBinds[0].OID(), varBinds[1])
		transport.mockGetNext("test", varBinds[1].OID(), varBinds[2])

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

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		var walkMock mock.Mock
		defer walkMock.AssertExpectations(t)

		transport.mockGetNextMulti("test", []snmp.OID{oid1, oid2}, []snmp.VarBind{varBinds[0], varBinds[0]})
		walkMock.On("walk[1/2]", errBind)
		walkMock.On("walk[2/2]", varBinds[0])
		transport.mockGetNextMulti("test", []snmp.OID{oid1, varBinds[0].OID()}, []snmp.VarBind{varBinds[0], varBinds[1]})
		walkMock.On("walk[1/2]", errBind)
		walkMock.On("walk[2/2]", varBinds[1])
		transport.mockGetNextMulti("test", []snmp.OID{oid1, varBinds[1].OID()}, []snmp.VarBind{varBinds[0], varBinds[2]})

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
