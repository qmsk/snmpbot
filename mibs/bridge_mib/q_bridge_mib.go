package bridge_mib

import (
	"github.com/qmsk/snmpbot/mibs"
)

var QMIB = mibs.RegisterMIB("Q-BRIDGE-MIB", 1, 3, 6, 1, 2, 1, 17, 7) // extends BRIDGE-QMIB

var (
	dot1qBase     = QMIB.MakeID("dot1qBase", 1, 1)
	dot1qTp       = QMIB.MakeID("dot1qTp", 1, 2)
	dot1qStatic   = QMIB.MakeID("dot1qStatic", 1, 3)
	dot1qVlan     = QMIB.MakeID("dot1qVlan", 1, 4)
	dot1vProtocol = QMIB.MakeID("dot1vProtocol", 1, 5)

	dot1qVlanVersionNumber = QMIB.RegisterObject(dot1qBase.MakeID("dot1qVlanVersionNumber", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	dot1qMaxVlanId = QMIB.RegisterObject(dot1qBase.MakeID("dot1qMaxVlanId", 2), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	dot1qMaxSupportedVlans = QMIB.RegisterObject(dot1qBase.MakeID("dot1qMaxSupportedVlans", 3), mibs.Object{
		Syntax: mibs.UnsignedSyntax{},
	})
	dot1qNumVlans = QMIB.RegisterObject(dot1qBase.MakeID("dot1qNumVlans", 4), mibs.Object{
		Syntax: mibs.UnsignedSyntax{},
	})
	dot1qGvrpStatus = QMIB.RegisterObject(dot1qBase.MakeID("dot1qGvrpStatus", 5), mibs.Object{
		Syntax: mibs.EnumSyntax{
			{1, "enabled"},
			{2, "disabled"},
		},
	})
)

var (
	dot1qFdbEntry = dot1qTp.MakeID("dot1qFdbEntry", 1, 1)

	dot1qFdbId = QMIB.RegisterObject(dot1qFdbEntry.MakeID("dot1qFdbId", 1), mibs.Object{
		Syntax: mibs.UnsignedSyntax{},
	})
	dot1qFdbIndexSyntax = mibs.IndexSyntax{
		dot1qFdbId,
	}

	dot1qFdbTable = QMIB.RegisterTable(dot1qTp.MakeID("dot1qFdbTable", 2), mibs.Table{
		IndexSyntax: dot1qFdbIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			dot1qFdbId,
			QMIB.RegisterObject(dot1qFdbEntry.MakeID("dot1qFdbDynamicCount", 2), mibs.Object{
				IndexSyntax: dot1qFdbIndexSyntax,
				Syntax:      mibs.CounterSyntax{},
			}),
		},
	})
)

var (
	dot1qTpFdbEntry = dot1qTp.MakeID("dot1qTpFdbEntry", 2, 1)

	dot1qTpFdbAddress = QMIB.RegisterObject(dot1qTpFdbEntry.MakeID("dot1qTpFdbAddress", 1), mibs.Object{
		Syntax: mibs.MACAddressSyntax{},
	})
	dot1qTpFdbIndexSyntax = mibs.IndexSyntax{
		dot1qFdbId,
		dot1qTpFdbAddress,
	}

	dot1qTpFdbTable = QMIB.RegisterTable(dot1qTp.MakeID("dot1qTpFdbTable", 2), mibs.Table{
		IndexSyntax: dot1qTpFdbIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			QMIB.RegisterObject(dot1qTpFdbEntry.MakeID("dot1qTpFdbPort", 2), mibs.Object{
				IndexSyntax: dot1qTpFdbIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			QMIB.RegisterObject(dot1qTpFdbEntry.MakeID("dot1qTpFdbStatus", 3), mibs.Object{
				IndexSyntax: dot1qTpFdbIndexSyntax,
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

var (
	dot1qVlanNumDeletes = QMIB.RegisterObject(dot1qVlan.MakeID("dot1qVlanNumDeletes", 1), mibs.Object{
		Syntax: mibs.CounterSyntax{},
	})

	dot1qVlanCurrentEntry = dot1qVlan.MakeID("dot1qVlanCurrentEntry", 2, 1)
	dot1qVlanTimeMark     = QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanTimeMark", 1), mibs.Object{
		Syntax: mibs.TimeTicksSyntax{},
	})
	dot1qVlanIndex = QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanIndex", 2), mibs.Object{
		Syntax: mibs.UnsignedSyntax{},
	})
	dot1qVlanCurrentIndexSyntax = mibs.IndexSyntax{
		dot1qVlanTimeMark,
		dot1qVlanIndex,
	}

	dot1qVlanCurrentTable = QMIB.RegisterTable(dot1qVlan.MakeID("dot1qVlanCurrentTable", 2), mibs.Table{
		IndexSyntax: dot1qVlanCurrentIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			// dot1qVlanTimeMark not-accessible .1
			// dot1qVlanIndex not-accessible .2
			QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanFdbId", 3), mibs.Object{
				IndexSyntax: dot1qVlanCurrentIndexSyntax,
				Syntax:      mibs.UnsignedSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanCurrentEgressPorts", 4), mibs.Object{
				IndexSyntax: dot1qVlanCurrentIndexSyntax,
				Syntax:      PortListSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanCurrentUntaggedPorts", 5), mibs.Object{
				IndexSyntax: dot1qVlanCurrentIndexSyntax,
				Syntax:      PortListSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanStatus", 6), mibs.Object{
				IndexSyntax: dot1qVlanCurrentIndexSyntax,
				Syntax: mibs.EnumSyntax{
					{1, "other"},
					{2, "permanent"},
					{3, "dynamicGvrp"},
				},
			}),
			QMIB.RegisterObject(dot1qVlanCurrentEntry.MakeID("dot1qVlanCreationTime", 7), mibs.Object{
				IndexSyntax: dot1qVlanCurrentIndexSyntax,
				Syntax:      mibs.TimeTicksSyntax{},
			}),
		},
	})
)

var (
	dot1qVlanStaticIndexSyntax = mibs.IndexSyntax{
		dot1qVlanIndex,
	}
	dot1qVlanStaticEntry = dot1qVlan.MakeID("dot1qVlanStaticEntry", 3, 1)
	dot1qVlanStaticTable = QMIB.RegisterTable(dot1qVlan.MakeID("dot1qVlanStaticTable", 3), mibs.Table{
		IndexSyntax: dot1qVlanStaticIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			QMIB.RegisterObject(dot1qVlanStaticEntry.MakeID("dot1qVlanStaticName", 1), mibs.Object{
				IndexSyntax: dot1qVlanStaticIndexSyntax,
				Syntax:      mibs.DisplayStringSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanStaticEntry.MakeID("dot1qVlanStaticEgressPorts", 2), mibs.Object{
				IndexSyntax: dot1qVlanStaticIndexSyntax,
				Syntax:      PortListSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanStaticEntry.MakeID("dot1qVlanForbiddenEgressPorts", 3), mibs.Object{
				IndexSyntax: dot1qVlanStaticIndexSyntax,
				Syntax:      PortListSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanStaticEntry.MakeID("dot1qVlanStaticUntaggedPorts", 4), mibs.Object{
				IndexSyntax: dot1qVlanStaticIndexSyntax,
				Syntax:      PortListSyntax{},
			}),
			QMIB.RegisterObject(dot1qVlanStaticEntry.MakeID("dot1qVlanStaticRowStatus", 5), mibs.Object{
				IndexSyntax: dot1qVlanStaticIndexSyntax,
				Syntax:      mibs.IntegerSyntax{}, // TODO: RowStatus
			}),
		},
	})

	dot1qNextFreeLocalVlanIndex = QMIB.RegisterObject(dot1qVlan.MakeID("dot1qNextFreeLocalVlanIndex", 4), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
)
