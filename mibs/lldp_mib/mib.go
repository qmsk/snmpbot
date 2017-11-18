package lldp_mib

import (
	"github.com/qmsk/snmpbot/mibs"
)

var MIB = mibs.RegisterMIB("LLDP-MIB", 1, 0, 8802, 1, 1, 2)

var (
	lldpObjects           = MIB.MakeID("lldpObjects", 1)
	lldpLocalSystemData   = lldpObjects.MakeID("lldpLocalSystemData", 3)
	lldpRemoteSystemsData = lldpObjects.MakeID("lldpRemoteSystemsData", 4)

	lldpChassisIdSubtypeSyntax = mibs.EnumSyntax{
		{1, "chassisComponent"},
		{2, "interfaceAlias"},
		{3, "portComponent"},
		{4, "macAddress"},
		{5, "networkAddress"},
		{6, "interfaceName"},
		{7, "local"},
	}
	lldpPortIdSubtypeSyntax = mibs.EnumSyntax{
		{1, "interfaceAlias"},
		{2, "portComponent"},
		{3, "macAddress"},
		{4, "networkAddress"},
		{5, "interfaceName"},
		{6, "agentCircuitId"},
		{7, "local"},
	}
	lldpPortIdSyntax = mibs.OctetStringSyntax{}

	lldpChassisIdSyntax  = mibs.OctetStringSyntax{}
	lldpPortNumberSyntax = mibs.IntegerSyntax{}
)

var (
	lldpLocChassisIdSubtype = MIB.RegisterObject(lldpLocalSystemData.MakeID("lldpLocChassisIdSubtype", 1), mibs.Object{
		Syntax: lldpChassisIdSubtypeSyntax,
	})
	lldpLocChassisId = MIB.RegisterObject(lldpLocalSystemData.MakeID("lldpLocChassisId", 2), mibs.Object{
		Syntax: lldpChassisIdSyntax,
	})
	lldpLocSysName = MIB.RegisterObject(lldpLocalSystemData.MakeID("lldpLocSysName", 3), mibs.Object{
		Syntax: mibs.DisplayStringSyntax{},
	})
	lldpLocSysDesc = MIB.RegisterObject(lldpLocalSystemData.MakeID("lldpLocSysDesc", 4), mibs.Object{
		Syntax: mibs.DisplayStringSyntax{},
	})
	// lldpLocSysCapSupported  // TODO: .5 BitsSyntax
	// lldpLocSysCapEnabled    // TODO: .6 BitsSyntax
)

var (
	lldpLocPortEntry = lldpLocalSystemData.MakeID("lldpLocPortEntry", 7, 1)
	lldpLocPortNum   = MIB.RegisterObject(lldpLocPortEntry.MakeID("lldpLocPortNum", 1), mibs.Object{
		Syntax:     lldpPortNumberSyntax,
		NotAccessible: true,
	})
	lldpLocPortIndexSyntax = mibs.IndexSyntax{
		lldpLocPortNum,
	}

	lldpLocPortTable = MIB.RegisterTable(lldpLocalSystemData.MakeID("lldpLocPortTable", 7), mibs.Table{
		IndexSyntax: lldpLocPortIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			// lldpLocPortNum       .1  not-accessible
			MIB.RegisterObject(lldpLocPortEntry.MakeID("lldpLocPortIdSubtype", 2), mibs.Object{
				IndexSyntax: lldpLocPortIndexSyntax,
				Syntax:      lldpPortIdSubtypeSyntax,
			}),
			MIB.RegisterObject(lldpLocPortEntry.MakeID("lldpLocPortId", 3), mibs.Object{
				IndexSyntax: lldpLocPortIndexSyntax,
				Syntax:      lldpPortIdSyntax,
			}),
			MIB.RegisterObject(lldpLocPortEntry.MakeID("lldpLocPortDesc", 4), mibs.Object{
				IndexSyntax: lldpLocPortIndexSyntax,
				Syntax:      mibs.DisplayStringSyntax{},
			}),
		},
	})
)

var (
	lldpRemEntry    = lldpRemoteSystemsData.MakeID("lldpRemEntry", 1, 1)
	lldpRemTimeMark = MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemTimeMark", 1), mibs.Object{
		Syntax:     mibs.TimeTicksSyntax{},
		NotAccessible: true,
	})
	lldpRemLocalPortNum = MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemLocalPortNum", 2), mibs.Object{
		Syntax:     lldpPortNumberSyntax,
		NotAccessible: true,
	})
	lldpRemIndex = MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemIndex", 3), mibs.Object{
		Syntax:     mibs.IntegerSyntax{},
		NotAccessible: true,
	})
	lldpRemIndexSyntax = mibs.IndexSyntax{
		lldpRemTimeMark,
		lldpRemLocalPortNum,
		lldpRemIndex,
	}

	lldpRemTable = MIB.RegisterTable(lldpRemoteSystemsData.MakeID("lldpRemTable", 1), mibs.Table{
		IndexSyntax: lldpRemIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemChassisIdSubtype", 4), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      lldpChassisIdSubtypeSyntax,
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemChassisId", 5), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      mibs.OctetStringSyntax{},
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemPortIdSubtype", 6), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      lldpPortIdSubtypeSyntax,
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemPortId", 7), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      mibs.OctetStringSyntax{},
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemPortDesc", 8), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      mibs.DisplayStringSyntax{},
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemSysName", 9), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      mibs.DisplayStringSyntax{},
			}),
			MIB.RegisterObject(lldpRemEntry.MakeID("lldpRemSysDesc", 10), mibs.Object{
				IndexSyntax: lldpRemIndexSyntax,
				Syntax:      mibs.DisplayStringSyntax{},
			}),
			// lldpRemSysCapSupported TODO: BitsSyntax
			// lldpRemSysCapEnabled TODO: BitsSyntax
		},
	})
)
