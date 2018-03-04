package client

import (
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
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
			PDU: snmp.GenericPDU{
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
			PDU: snmp.GenericPDU{
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
			PDU: snmp.GenericPDU{
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
			PDU: snmp.GenericPDU{
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
			PDU: snmp.GenericPDU{
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

func TestGetRequestTooBig(t *testing.T) {
	var oids = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 0},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 3},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 4},
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.On("GetRequest", IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetRequestType,
			PDU: snmp.GenericPDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(oids[0], nil),
					snmp.MakeVarBind(oids[1], nil),
					snmp.MakeVarBind(oids[2], nil),
					snmp.MakeVarBind(oids[3], nil),
					snmp.MakeVarBind(oids[4], nil),
				},
			},
		}).Return(error(nil), IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.GenericPDU{
				ErrorStatus: snmp.TooBigError,
			},
		})

		if varBinds, err := client.Get(oids...); err == nil {
			t.Errorf("Get(%v): %v", oids, varBinds)
		} else {
			assert.EqualError(t, err, "SNMP GetRequest error: SNMP PDU Error: TooBig @ .")
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

func TestGetRequestGetBulk(t *testing.T) {
	var scalarOID = snmp.OID{1, 3, 6, 1, 2, 1, 1, 4}
	var entryOIDs = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 20
		client.options.MaxRepetitions = 5

		transport.On("GetBulkRequest", IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetBulkRequestType,
			PDU: snmp.BulkPDU{
				NonRepeaters:   1,
				MaxRepetitions: 5,
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, nil),
					snmp.MakeVarBind(entryOIDs[0], nil),
					snmp.MakeVarBind(entryOIDs[1], nil),
				},
			},
		}).Return(error(nil), IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.GenericPDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, 0),
					snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
					snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
					snmp.MakeVarBind(entryOIDs[0].Extend(2), 3),
					snmp.MakeVarBind(entryOIDs[1].Extend(2), 4),
				},
			},
		})

		scalarVars, entryVarsList, err := client.GetBulk([]snmp.OID{scalarOID}, entryOIDs)
		if err != nil {
			t.Errorf("GetBulk(%v, %v): %v", scalarOID, entryOIDs, err)
		}

		assert.Equal(t, []snmp.VarBind{snmp.MakeVarBind(scalarOID, 0)}, scalarVars)
		assert.Equal(t, [][]snmp.VarBind{
			[]snmp.VarBind{
				snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
				snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
			},
			[]snmp.VarBind{
				snmp.MakeVarBind(entryOIDs[0].Extend(2), 3),
				snmp.MakeVarBind(entryOIDs[1].Extend(2), 4),
			},
		}, entryVarsList)
	})
}

func TestGetRequestGetBulkOne(t *testing.T) {
	var scalarOID = snmp.OID{1, 3, 6, 1, 2, 1, 1, 4}
	var entryOIDs = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 20
		client.options.MaxRepetitions = 5

		transport.On("GetBulkRequest", IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetBulkRequestType,
			PDU: snmp.BulkPDU{
				NonRepeaters:   1,
				MaxRepetitions: 5,
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, nil),
					snmp.MakeVarBind(entryOIDs[0], nil),
					snmp.MakeVarBind(entryOIDs[1], nil),
				},
			},
		}).Return(error(nil), IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.GenericPDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, 0),
					snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
					snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
				},
			},
		})

		scalarVars, entryVarsList, err := client.GetBulk([]snmp.OID{scalarOID}, entryOIDs)
		if err != nil {
			t.Errorf("GetBulk(%v, %v): %v", scalarOID, entryOIDs, err)
		}

		assert.Equal(t, []snmp.VarBind{snmp.MakeVarBind(scalarOID, 0)}, scalarVars)
		assert.Equal(t, [][]snmp.VarBind{
			[]snmp.VarBind{
				snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
				snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
			},
		}, entryVarsList)
	})
}

func TestGetRequestGetBulkShort(t *testing.T) {
	var scalarOID = snmp.OID{1, 3, 6, 1, 2, 1, 1, 4}
	var entryOIDs = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 20
		client.options.MaxRepetitions = 5

		transport.On("GetBulkRequest", IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetBulkRequestType,
			PDU: snmp.BulkPDU{
				NonRepeaters:   1,
				MaxRepetitions: 5,
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, nil),
					snmp.MakeVarBind(entryOIDs[0], nil),
					snmp.MakeVarBind(entryOIDs[1], nil),
				},
			},
		}).Return(error(nil), IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.GenericPDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, 0),
				},
			},
		})

		_, _, err := client.GetBulk([]snmp.OID{scalarOID}, entryOIDs)

		assert.EqualError(t, err, "Invalid bulk response for 1+2 => 1 vars")
	})
}

func TestGetRequestGetBulkOdd(t *testing.T) {
	var scalarOID = snmp.OID{1, 3, 6, 1, 2, 1, 1, 4}
	var entryOIDs = []snmp.OID{
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
		snmp.OID{1, 3, 6, 1, 2, 1, 1, 5, 2},
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 5
		client.options.MaxRepetitions = 5

		transport.On("GetBulkRequest", IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetBulkRequestType,
			PDU: snmp.BulkPDU{
				NonRepeaters:   1,
				MaxRepetitions: 2,
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, nil),
					snmp.MakeVarBind(entryOIDs[0], nil),
					snmp.MakeVarBind(entryOIDs[1], nil),
				},
			},
		}).Return(error(nil), IO{
			Addr: testAddr("test"),
			Packet: snmp.Packet{
				Version:   snmp.SNMPv2c,
				Community: []byte("public"),
			},
			PDUType: snmp.GetResponseType,
			PDU: snmp.GenericPDU{
				VarBinds: []snmp.VarBind{
					snmp.MakeVarBind(scalarOID, 0),
					snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
					snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
					snmp.MakeVarBind(entryOIDs[0].Extend(1), 3),
				},
			},
		})

		scalarVars, entryVarsList, err := client.GetBulk([]snmp.OID{scalarOID}, entryOIDs)
		if err != nil {
			t.Errorf("GetBulk(%v, %v): %v", scalarOID, entryOIDs, err)
		}

		assert.Equal(t, []snmp.VarBind{snmp.MakeVarBind(scalarOID, 0)}, scalarVars)
		assert.Equal(t, [][]snmp.VarBind{
			[]snmp.VarBind{
				snmp.MakeVarBind(entryOIDs[0].Extend(1), 1),
				snmp.MakeVarBind(entryOIDs[1].Extend(1), 2),
			},
		}, entryVarsList)
	})
}
