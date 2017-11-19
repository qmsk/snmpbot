// SNMP BRIDGE-MIB implementation
package bridge_mib

import (
	"github.com/qmsk/snmpbot/mibs"
)

var MIB = mibs.RegisterMIB("BRIDGE-MIB", 1, 3, 6, 1, 2, 1, 17)

var (
	dot1dBase = MIB.MakeID("dot1dBase", 1)
	dot1dStp  = MIB.MakeID("dot1dStp", 2)
	dot1dTp   = MIB.MakeID("dot1dTp", 4)
)

var (
	dot1dBaseBridgeAddress = MIB.RegisterObject(dot1dBase.MakeID("dot1dBaseBridgeAddress", 1), mibs.Object{
		Syntax: mibs.MACAddressSyntax{},
	})
	dot1dBaseNumPorts = MIB.RegisterObject(dot1dBase.MakeID("dot1dBaseNumPorts", 2), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	dot1dBaseType = MIB.RegisterObject(dot1dBase.MakeID("dot1dBaseType", 3), mibs.Object{
		Syntax: mibs.EnumSyntax{
			{1, "unknown"},
			{2, "transparent-only"},
			{3, "sourceroute-only"},
			{4, "srt"},
		},
	})
)
var (
	dot1dStpProtocolSpecification = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpProtocolSpecification", 1), mibs.Object{
		Syntax: mibs.EnumSyntax{
			{1, "unknown"},
			{2, "decLb100"},
			{3, "ieee8021d"},
		},
	})
	dot1dStpPriority = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpPriority", 2), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	dot1dStpTimeSinceTopologyChange = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpTimeSinceTopologyChange", 3), mibs.Object{
		Syntax: mibs.TimeTicksSyntax{},
	})
	dot1dStpTopChanges = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpTopChanges", 4), mibs.Object{
		Syntax: mibs.CounterSyntax{},
	})
	dot1dStpDesignatedRoot = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpDesignatedRoot", 5), mibs.Object{
		Syntax: BridgeIDSyntax{},
	})
	dot1dStpRootCost = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpRootCost", 6), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	dot1dStpRootPort = MIB.RegisterObject(dot1dStp.MakeID("dot1dStpRootPort", 7), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
)

var (
	dot1dStpPortTableID = dot1dStp.MakeID("dot1dStpPortTable", 15)
	dot1dStpPortEntry   = dot1dStpPortTableID.MakeID("dot1dStpPortTable", 1)

	dot1dStpPort = MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPort", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})

	dot1dStpPortIndexSyntax = mibs.IndexSyntax{
		dot1dStpPort,
	}
	dot1dStpPortTable = MIB.RegisterTable(dot1dStpPortEntry, mibs.Table{
		IndexSyntax: dot1dStpPortIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			dot1dStpPort,
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortPriority", 2), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortState", 3), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax: mibs.EnumSyntax{
					{1, "disabled"},
					{2, "blocking"},
					{3, "listening"},
					{4, "learning"},
					{5, "forwarding"},
					{6, "broken"},
				},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortEnable", 4), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax: mibs.EnumSyntax{
					{1, "enabled"},
					{2, "disabled"},
				},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortPathCost", 5), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortDesignatedRoot", 6), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      BridgeIDSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortDesignatedCost", 7), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortDesignatedBridge", 8), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      BridgeIDSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortDesignatedPort", 9), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      mibs.OctetStringSyntax{},
			}),
			MIB.RegisterObject(dot1dStpPortEntry.MakeID("dot1dStpPortForwardTransitions", 10), mibs.Object{
				IndexSyntax: dot1dStpPortIndexSyntax,
				Syntax:      mibs.CounterSyntax{},
			}),
			// dot1dStpPortPathCost32 XXX: compat issues
		},
	})
)

var (
	dot1dTpLearnedEntryDiscards = MIB.RegisterObject(dot1dTp.MakeID("dot1dTpLearnedEntryDiscards", 1), mibs.Object{
		Syntax: mibs.CounterSyntax{},
	})
	dot1dTpAgingTime = MIB.RegisterObject(dot1dTp.MakeID("dot1dTpAgingTime", 2), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})

	dot1dTpFdbEntry   = dot1dTp.MakeID("dot1dTpFdbEntry", 3, 1)
	dot1dTpFdbAddress = MIB.RegisterObject(dot1dTpFdbEntry.MakeID("dot1dTpFdbAddress", 1), mibs.Object{
		Syntax: mibs.MACAddressSyntax{},
	})
	dot1dTpFdbIndexSyntax = mibs.IndexSyntax{
		dot1dTpFdbAddress,
	}

	dot1dTpFdbTable = MIB.RegisterTable(dot1dTp.MakeID("dot1dTpFdbTable", 3), mibs.Table{
		IndexSyntax: dot1dTpFdbIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			dot1dTpFdbAddress,
			MIB.RegisterObject(dot1dTpFdbEntry.MakeID("dot1dTpFdbPort", 2), mibs.Object{
				IndexSyntax: dot1dTpFdbIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			MIB.RegisterObject(dot1dTpFdbEntry.MakeID("dot1dTpFdbStatus", 3), mibs.Object{
				IndexSyntax: dot1dTpFdbIndexSyntax,
				Syntax: mibs.EnumSyntax{
					{1, "other"},
					{2, "invalid"},
					{3, "learned"},
					{4, "self"},
					{5, "mgmt"},
				},
			}),
		},
	})
)

func init() {
	dot1dStpPort.IndexSyntax = dot1dStpPortIndexSyntax
	dot1dTpFdbAddress.IndexSyntax = dot1dTpFdbIndexSyntax
}
