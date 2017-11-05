package if_mib

import (
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
)

// SNMP IF-MIB implementation

var MIB = mibs.RegisterMIB("IF-MIB", snmp.OID{1, 3, 6, 1, 2, 1, 31})
var Interfaces = mibs.RegisterMIB("interfaces", snmp.OID{1, 3, 6, 1, 2, 1, 2})

var (
	Number = Interfaces.RegisterObject(Interfaces.MakeID("ifNumber", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax,
	})
	TableLastChange = MIB.RegisterObject(Interfaces.MakeID("ifTableLastChange", 1, 5), mibs.Object{
		Syntax: mibs.IntegerSyntax,
	})
	Table = Interfaces.RegisterTable(Interfaces.MakeID("ifTable", 2), mibs.Table{
		Index: []mibs.TableIndex{
			{"ifIndex", mibs.IntegerIndexSyntax},
		},
	})

	Index = Table.RegisterObject(Table.MakeID("ifIndex", 1, 1), mibs.Object{
		Syntax: mibs.IntegerSyntax,
	})
	Descr = Table.RegisterObject(Table.MakeID("ifDescr", 1, 2), mibs.Object{
		Syntax: mibs.DisplayStringSyntax,
	})
	Type = Table.RegisterObject(Table.MakeID("ifType", 1, 3), mibs.Object{
		Syntax: mibs.IntegerSyntax,
	})
	Mtu = Table.RegisterObject(Table.MakeID("ifMtu", 1, 4), mibs.Object{
		Syntax: mibs.IntegerSyntax,
	})
	Speed = Table.RegisterObject(Table.MakeID("ifSpeed", 1, 5), mibs.Object{
		Syntax: mibs.GaugeSyntax,
	})
	PhysAddress = Table.RegisterObject(Table.MakeID("ifPhysAddress", 1, 6), mibs.Object{
		Syntax: mibs.PhysAddressSyntax,
	})
	AdminStatus = Table.RegisterObject(Table.MakeID("ifAdminStatus", 1, 7), mibs.Object{
		Syntax: mibs.EnumSyntax{
			{1, "up"},
			{2, "down"},
			{3, "testing"},
		},
	})
	OperStatus = Table.RegisterObject(Table.MakeID("ifOperStatus", 1, 8), mibs.Object{
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
	LastChange = Table.RegisterObject(Table.MakeID("ifLastChange", 1, 9), mibs.Object{
		Syntax: mibs.TimeTicksSyntax,
	})

/*
	Interfaces_linkDown = SNMPv2MIB.registerNotificationType("linkDown", Interfaces.define(1, 5, 3))
	Interfaces_linkUp   = SNMPv2MIB.registerNotificationType("linkUp", Interfaces.define(1, 5, 4))
*/
)
