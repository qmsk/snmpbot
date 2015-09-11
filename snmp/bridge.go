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
        Index:  TableIndex{Name: "dot1dTpFdbAddress", IndexSyntax: MacAddressSyntax},
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
