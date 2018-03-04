package client

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/mock"
	"testing"
)

type walkResult struct {
	scalars []snmp.VarBind
	entries []snmp.VarBind
}

type walkTest struct {
	scalars []snmp.OID
	entries []snmp.OID

	results []walkResult
}

func testWalk(t *testing.T, client *Client, test walkTest) {
	var walkMock mock.Mock
	defer walkMock.AssertExpectations(t)

	for _, result := range test.results {
		if result.scalars == nil {
			result.scalars = []snmp.VarBind{}
		}
		if result.entries == nil {
			result.entries = []snmp.VarBind{}
		}
		walkMock.On("walk", result.scalars, result.entries).Once()
	}

	if err := client.WalkWithScalars(test.scalars, test.entries, func(scalars []snmp.VarBind, entries []snmp.VarBind) error {
		walkMock.MethodCalled("walk", scalars, entries)

		return nil
	}); err != nil {
		t.Fatalf("Walk(%v, %v): %v", test.scalars, test.entries, err)
	}
}

func TestWalkTable(t *testing.T) {
	var ifName = snmp.MustParseOID(".1.3.6.1.2.1.31.1.1.1.1")            // IF-MIB::ifName
	var ifInMulticastPkts = snmp.MustParseOID(".1.3.6.1.2.1.31.1.1.1.2") // IF-MIB::ifInMulticastPkts

	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(ifName.Extend(1), []byte("if1")),
		snmp.MakeVarBind(ifName.Extend(2), []byte("if2")),
		snmp.MakeVarBind(ifInMulticastPkts.Extend(0), snmp.Counter32(0)),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNext("test", ifName, varBinds[0])
		transport.mockGetNext("test", ifName.Extend(1), varBinds[1])
		transport.mockGetNext("test", ifName.Extend(2), varBinds[2])

		testWalk(t, client, walkTest{
			entries: []snmp.OID{ifName},
			results: []walkResult{
				{entries: []snmp.VarBind{varBinds[0]}},
				{entries: []snmp.VarBind{varBinds[1]}},
			},
		})
	})
}

func TestWalkScalars(t *testing.T) {
	var oid = snmp.MustParseOID(".1.3.6.1.2.1.2.1") // IF-MIB::ifNumber
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(oid.Extend(0), int(2)),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNext("test", oid, varBinds[0])

		testWalk(t, client, walkTest{
			scalars: []snmp.OID{oid},
			results: []walkResult{
				{scalars: []snmp.VarBind{varBinds[0]}},
			},
		})
	})
}

func TestWalk(t *testing.T) {
	var ifNumber = snmp.MustParseOID(".1.3.6.1.2.1.2.1")                 // IF-MIB::ifNumber
	var ifName = snmp.MustParseOID(".1.3.6.1.2.1.31.1.1.1.1")            // IF-MIB::ifName
	var ifInMulticastPkts = snmp.MustParseOID(".1.3.6.1.2.1.31.1.1.1.2") // IF-MIB::ifInMulticastPkts

	var varBind = snmp.MakeVarBind(ifNumber.Extend(0), int(2))
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(ifName.Extend(1), []byte("if1")),
		snmp.MakeVarBind(ifName.Extend(2), []byte("if2")),
		snmp.MakeVarBind(ifInMulticastPkts, snmp.Counter32(0)),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNextMulti("test", []snmp.OID{ifNumber, ifName}, []snmp.VarBind{varBind, varBinds[0]})
		transport.mockGetNextMulti("test", []snmp.OID{ifNumber, ifName.Extend(1)}, []snmp.VarBind{varBind, varBinds[1]})
		transport.mockGetNextMulti("test", []snmp.OID{ifNumber, ifName.Extend(2)}, []snmp.VarBind{varBind, varBinds[2]})

		testWalk(t, client, walkTest{
			scalars: []snmp.OID{ifNumber},
			entries: []snmp.OID{ifName},
			results: []walkResult{
				{scalars: []snmp.VarBind{varBind}, entries: []snmp.VarBind{varBinds[0]}},
				{scalars: []snmp.VarBind{varBind}, entries: []snmp.VarBind{varBinds[1]}},
			},
		})
	})
}

func TestWalkV2(t *testing.T) {
	var oid = snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1} // IF-MIB::ifName
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 1}, []byte("if1")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 2}, []byte("if2")),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 31, 1, 1, 1, 1, 2}, snmp.EndOfMibViewValue),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNext("test", oid, varBinds[0])
		transport.mockGetNext("test", varBinds[0].OID(), varBinds[1])
		transport.mockGetNext("test", varBinds[1].OID(), varBinds[2])

		testWalk(t, client, walkTest{
			entries: []snmp.OID{oid},
			results: []walkResult{
				{entries: []snmp.VarBind{varBinds[0]}},
				{entries: []snmp.VarBind{varBinds[1]}},
			},
		})
	})
}

func TestWalkTablePartial(t *testing.T) {
	var oid1 = snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 1} // Q-BRIDGE-MIB::dot1qTpFdbAddress (not-accessible)
	var oid2 = snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2} // Q-BRIDGE-MIB::dot1qTpFdbPort

	var errBind = snmp.MakeVarBind(oid1, snmp.EndOfMibViewValue)
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2, 1, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}, int(1)),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 2, 1, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}, int(3)),
		snmp.MakeVarBind(snmp.OID{1, 3, 6, 1, 2, 1, 17, 7, 1, 2, 2, 1, 3, 1, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}, int(1)),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		transport.mockGetNextMulti("test", []snmp.OID{oid1, oid2}, []snmp.VarBind{varBinds[0], varBinds[0]})
		transport.mockGetNextMulti("test", []snmp.OID{oid1, varBinds[0].OID()}, []snmp.VarBind{varBinds[0], varBinds[1]})
		transport.mockGetNextMulti("test", []snmp.OID{oid1, varBinds[1].OID()}, []snmp.VarBind{varBinds[0], varBinds[2]})

		testWalk(t, client, walkTest{
			entries: []snmp.OID{oid1, oid2},
			results: []walkResult{
				{entries: []snmp.VarBind{errBind, varBinds[0]}},
				{entries: []snmp.VarBind{errBind, varBinds[1]}},
			},
		})
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
	var varBinds = []snmp.VarBind{
		snmp.MakeVarBind(oids[0].Extend(0), []byte("qmsk-snmp test 0")),
		snmp.MakeVarBind(oids[1].Extend(0), []byte("qmsk-snmp test 1")),
		snmp.MakeVarBind(oids[2].Extend(0), []byte("qmsk-snmp test 2")),
		snmp.MakeVarBind(oids[3].Extend(0), []byte("qmsk-snmp test 3")),
		snmp.MakeVarBind(oids[4].Extend(0), []byte("qmsk-snmp test 4")),
	}

	withTestClient(t, "test", func(transport *testTransport, client *Client) {
		client.options.MaxVars = 2

		transport.mockGetNextMulti("test", []snmp.OID{oids[0], oids[1]}, []snmp.VarBind{
			varBinds[0],
			varBinds[1],
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[2], oids[3]}, []snmp.VarBind{
			varBinds[2],
			varBinds[3],
		})
		transport.mockGetNextMulti("test", []snmp.OID{oids[4]}, []snmp.VarBind{
			varBinds[4],
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

		testWalk(t, client, walkTest{
			entries: oids,
			results: []walkResult{
				{entries: varBinds},
			},
		})
	})
}
