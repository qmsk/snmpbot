package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/mock"
	"testing"
)

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

func TestWalkMany(t *testing.T) {
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

		transport.mockGetNextMulti("test", []snmp.OID{oids[0], oids[1]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[0].Extend(0), values[0]),
			snmp.MakeVarBind(oids[1].Extend(0), values[1]),
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[2], oids[3]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[2].Extend(0), values[2]),
			snmp.MakeVarBind(oids[3].Extend(0), values[3]),
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[4]}, []snmp.VarBind{
			snmp.MakeVarBind(oids[4].Extend(0), values[4]),
		})

		transport.mockGetNextMulti("test", []snmp.OID{oids[0].Extend(0), oids[1].Extend(0)}, []snmp.VarBind{
			snmp.MakeVarBind(oids[0].Extend(0), snmp.EndOfMibViewValue),
			snmp.MakeVarBind(oids[1].Extend(0), snmp.EndOfMibViewValue),
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[2].Extend(0), oids[3].Extend(0)}, []snmp.VarBind{
			snmp.MakeVarBind(oids[2].Extend(0), snmp.EndOfMibViewValue),
			snmp.MakeVarBind(oids[3].Extend(0), snmp.EndOfMibViewValue),
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[4].Extend(0)}, []snmp.VarBind{
			snmp.MakeVarBind(oids[4].Extend(0), snmp.EndOfMibViewValue),
		})

		var walkMock mock.Mock
		defer walkMock.AssertExpectations(t)

		walkMock.On("walk", []snmp.VarBind{
			snmp.MakeVarBind(oids[0].Extend(0), values[0]),
			snmp.MakeVarBind(oids[1].Extend(0), values[1]),
			snmp.MakeVarBind(oids[2].Extend(0), values[2]),
			snmp.MakeVarBind(oids[3].Extend(0), values[3]),
			snmp.MakeVarBind(oids[4].Extend(0), values[4]),
		})

		if err := client.Walk(func(varBinds ...snmp.VarBind) error {
			walkMock.MethodCalled("walk", varBinds)

			return nil
		}, oids...); err != nil {
			t.Fatalf("Walk(%v): %v", oids, err)
		}
	})
}
