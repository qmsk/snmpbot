package snmp

// SNMP BRIDGE-MIB implementation
var (
    BridgeMIB       = registerMIB("BRIDGE-MIB", OID{1,3,6,1,2,1,17})

    Bridge_dot1dBase    = BridgeMIB.define(1)
    Bridge_dot1dTp      = BridgeMIB.define(4)

    Bridge_dot1dBaseBridgeAddress       = BridgeMIB.registerObject("dot1dBaseBridgeAddress",        MacAddressSyntax,   Bridge_dot1dBase.define(1))
    Bridge_dot1dBaseNumPorts            = BridgeMIB.registerObject("dot1dBaseNumPorts",             IntegerSyntax,      Bridge_dot1dBase.define(2))
    Bridge_dot1dBaseType                = BridgeMIB.registerObject("dot1dBaseType",                 IntegerSyntax,      Bridge_dot1dBase.define(3))

    Bridge_dot1dTpLearnedEntryDiscards  = BridgeMIB.registerObject("dot1dTpLearnedEntryDiscards",   CounterSyntax,      Bridge_dot1dTp.define(1))
    Bridge_dot1dTpAgingTime             = BridgeMIB.registerObject("dot1dTpAgingTime",              IntegerSyntax,      Bridge_dot1dTp.define(2))

    Bridge_dot1dTpFdbEntry  = Bridge_dot1dTp.define(3, 1)
    Bridge_dot1dTpFdbTable          = BridgeMIB.registerTable(&Table{Node:Node{OID: Bridge_dot1dTp.define(3), Name: "dot1dTpFdbTable"},
        Index:  []TableIndex{
            {"dot1dTpFdbAddress", MacAddressSyntax},
        },
        Entry: []*Object{
            BridgeMIB.registerObject("dot1dTpFdbAddress",   MacAddressSyntax,   Bridge_dot1dTpFdbEntry.define(1)),
            BridgeMIB.registerObject("dot1dTpFdbPort",      IntegerSyntax,      Bridge_dot1dTpFdbEntry.define(2)),
            BridgeMIB.registerObject("dot1dTpFdbStatus",    EnumSyntax{
                {1, "other"},
                {2, "invalid"},
                {3, "learned"},
                {4, "self"},
                {5, "mgmt"},
            }, Bridge_dot1dTpFdbEntry.define(3)),
        },
    })
)

var (
    P_BRIDGE_EnabledStatusSyntax        = EnumSyntax{
        {1, "enabled"},
        {2, "disabled"},
    }
)

var (
    Q_BRIDGE_MIB        = registerMIB("Q-BRIDGE-MIB", OID{1,3,6,1,2,1,17,7})    // extends BRIDGE-MIB

    Q_BRIDGE_dot1qBase      = Q_BRIDGE_MIB.define(1,1)
    Q_BRIDGE_dot1qTp        = Q_BRIDGE_MIB.define(1,2)
    Q_BRIDGE_dot1qStatic    = Q_BRIDGE_MIB.define(1,3)
    Q_BRIDGE_dot1qVlan      = Q_BRIDGE_MIB.define(1,4)
    Q_BRIDGE_dot1vProtocol  = Q_BRIDGE_MIB.define(1,5)

    Q_BRIDGE_dot1qVlanVersionNumber = Q_BRIDGE_MIB.registerObject("dot1qVlanVersionNumber", IntegerSyntax,  Q_BRIDGE_dot1qBase.define(1))
    Q_BRIDGE_dot1qMaxVlanId         = Q_BRIDGE_MIB.registerObject("dot1qMaxVlanId",         IntegerSyntax,  Q_BRIDGE_dot1qBase.define(2))
    Q_BRIDGE_dot1qMaxSupportedVlans = Q_BRIDGE_MIB.registerObject("dot1qMaxSupportedVlans", UnsignedSyntax, Q_BRIDGE_dot1qBase.define(3))
    Q_BRIDGE_dot1qNumVlans          = Q_BRIDGE_MIB.registerObject("dot1qNumVlans",          UnsignedSyntax, Q_BRIDGE_dot1qBase.define(4))
    Q_BRIDGE_dot1qGvrpStatus        = Q_BRIDGE_MIB.registerObject("dot1qGvrpStatus",        P_BRIDGE_EnabledStatusSyntax,   Q_BRIDGE_dot1qBase.define(5))

    Q_BRIDGE_dot1qTpFdbEntry        = Q_BRIDGE_dot1qTp.define(2, 1)
    Q_BRIDGE_dot1qTpFdbTable        = Q_BRIDGE_MIB.registerTable(&Table{Node:Node{OID: Q_BRIDGE_dot1qTp.define(2), Name: "dot1qTpFdbTable"},
        Index:  []TableIndex{
            {"dot1qFdbId",          UnsignedSyntax},
            {"dot1qTpFdbAddress",   MacAddressSyntax},
        },
        Entry:  []*Object{
            // dot1qTpFdbAddress .1 not-accessible
            Q_BRIDGE_MIB.registerObject("dot1qTpFdbPort",      IntegerSyntax,          Q_BRIDGE_dot1qTpFdbEntry.define(2)),
            Q_BRIDGE_MIB.registerObject("dot1qTpFdbStatus",    EnumSyntax{
                {1, "other"},
                {2, "invalid"},
                {3, "learned"},
                {4, "self"},
                {5, "mgmt"},
            }, Q_BRIDGE_dot1qTpFdbEntry.define(3)),
        },
    })

    Q_BRIDGE_dot1qVlanNumDeletes    = Q_BRIDGE_MIB.registerObject("dot1qVlanNumDeletes", CounterSyntax, Q_BRIDGE_dot1qVlan.define(1))
    Q_BRIDGE_dot1qVlanCurrentEntry  = Q_BRIDGE_dot1qVlan.define(2, 1)
    Q_BRIDGE_dot1qVlanCurrentTable  = Q_BRIDGE_MIB.registerTable(&Table{Node:Node{OID: Q_BRIDGE_dot1qVlan.define(2), Name: "dot1qVlanCurrentTable"},
        Index: []TableIndex{
            {"dot1qVlanTimeMark",   TimeTicksSyntax},
            {"dot1qVlanIndex",      UnsignedSyntax},
        },
        Entry:  []*Object{
            // dot1qVlanTimeMark not-accessible .1
            // dot1qVlanIndex not-accessible .2
            Q_BRIDGE_MIB.registerObject("dot1qVlanFdbId",                   UnsignedSyntax,     Q_BRIDGE_dot1qVlanCurrentEntry.define(3)),
            Q_BRIDGE_MIB.registerObject("dot1qVlanCurrentEgressPorts",      BinarySyntax,       Q_BRIDGE_dot1qVlanCurrentEntry.define(4)),
            Q_BRIDGE_MIB.registerObject("dot1qVlanCurrentUntaggedPorts",    BinarySyntax,       Q_BRIDGE_dot1qVlanCurrentEntry.define(5)),
            Q_BRIDGE_MIB.registerObject("dot1qVlanStatus",                  EnumSyntax{
                {1, "other"},
                {2, "permanent"},
                {3, "dynamicGvrp"},
            }, Q_BRIDGE_dot1qVlanCurrentEntry.define(6)),
            Q_BRIDGE_MIB.registerObject("dot1qVlanCreationTime",            TimeTicksSyntax,    Q_BRIDGE_dot1qVlanCurrentEntry.define(7)),
        },
    })
)

type Bridge_FdbIndex struct {
    Address     MacAddress
}

func (self *Bridge_FdbIndex) parseIndex (oid OID) (interface{}, error) {
    if address, err := self.Address.parseIndex(oid); err != nil {
        return nil, err
    } else {
        return Bridge_FdbIndex{Address: address.(MacAddress)}, nil
    }
}

func (self Bridge_FdbIndex) String() string {
    return self.Address.String()
}

type Bridge_FdbEntry struct {
    Address     MacAddress  `snmp:"1.3.6.1.2.1.17.4.3.1.1"`
    Port        Integer     `snmp:"1.3.6.1.2.1.17.4.3.1.2"`
    Status      Integer     `snmp:"1.3.6.1.2.1.17.4.3.1.3"`
}

type Bridge_FdbTable map[Bridge_FdbIndex]*Bridge_FdbEntry
