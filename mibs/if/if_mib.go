package if_mib

import (
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
)

// SNMP IF-MIB implementation

var MIB = mibs.RegisterMIB("IF-MIB", snmp.OID{1, 3, 6, 1, 2, 1, 31})
var InterfacesMIB = mibs.RegisterMIB("interfaces", snmp.OID{1, 3, 6, 1, 2, 1, 2})

var (
	Number = InterfacesMIB.RegisterObject(InterfacesMIB.MakeID("ifNumber", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	TableLastChange = MIB.RegisterObject(MIB.MakeID("ifTableLastChange", 1, 5), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	TableID = InterfacesMIB.MakeID("ifTable", 2)
	EntryID = TableID.MakeID("ifEntry", 1)

	Index = InterfacesMIB.RegisterObject(EntryID.MakeID("ifIndex", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})

	IndexSyntax = mibs.IndexSyntax{
		Index,
	}

	Descr = InterfacesMIB.RegisterObject(EntryID.MakeID("ifDescr", 2), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.DisplayStringSyntax{},
	})
	Type = InterfacesMIB.RegisterObject(EntryID.MakeID("ifType", 3), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.IntegerSyntax{},
	})
	Mtu = InterfacesMIB.RegisterObject(EntryID.MakeID("ifMtu", 4), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.IntegerSyntax{},
	})
	Speed = InterfacesMIB.RegisterObject(EntryID.MakeID("ifSpeed", 5), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.GaugeSyntax{},
	})
	PhysAddress = InterfacesMIB.RegisterObject(EntryID.MakeID("ifPhysAddress", 6), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.PhysAddressSyntax{},
	})
	AdminStatus = InterfacesMIB.RegisterObject(EntryID.MakeID("ifAdminStatus", 7), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.EnumSyntax{
			{1, "up"},
			{2, "down"},
			{3, "testing"},
		},
	})
	OperStatus = InterfacesMIB.RegisterObject(EntryID.MakeID("ifOperStatus", 8), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.EnumSyntax{
			{1, "up"},
			{2, "down"},
			{3, "testing"},
			{4, "unknown"},
			{5, "dormant"},
			{6, "notPresent"},
			{7, "lowerLayerDown"},
		},
	})
	LastChange = InterfacesMIB.RegisterObject(EntryID.MakeID("ifLastChange", 9), mibs.Object{
		IndexSyntax: IndexSyntax,
		Syntax: mibs.TimeTicksSyntax{},
	})

	Table = InterfacesMIB.RegisterTable(TableID, mibs.Table{
		IndexSyntax: IndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			Index,
			Descr,
			Type,
			Mtu,
			Speed,
			PhysAddress,
			AdminStatus,
			OperStatus,
			LastChange,
		},
	})

/*
	Interfaces_linkDown = SNMPv2MIB.registerNotificationType("linkDown", Interfaces.define(1, 5, 3))
	Interfaces_linkUp   = SNMPv2MIB.registerNotificationType("linkUp", Interfaces.define(1, 5, 4))
*/
)
