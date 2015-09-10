package snmp

// SNMP IF-MIB implementation

var (
    IfMIB       = registerMIB("IF-MIB",     1,3,6,1,2,1,31)
    Interfaces  = registerMIB("interfaces", 1,3,6,1,2,1,2)

    If_linkDown = SNMPv2MIB.registerNotificationType("linkDown", 1, 5, 3)
    If_linkUp   = SNMPv2MIB.registerNotificationType("linkUp", 1, 5, 4)
)

type InterfaceIndex struct {
    Index       Integer
}

func (self *InterfaceIndex) setIndex (oid OID) error {
    return self.Index.setIndex(oid)
}

func (self InterfaceIndex) String() string {
    return self.Index.String()
}

type InterfaceEntry struct {
    Index       Integer     `snmp:"1.3.6.1.2.1.2.2.1.1"`
    Descr       String      `snmp:"1.3.6.1.2.1.2.2.1.2"`
    Type        Integer     `snmp:"1.3.6.1.2.1.2.2.1.3"`
    Mtu         Integer     `snmp:"1.3.6.1.2.1.2.2.1.4"`
    Speed       Gauge       `snmp:"1.3.6.1.2.1.2.2.1.5"`
    PhysAddress Binary      `snmp:"1.3.6.1.2.1.2.2.1.6"`
    AdminStatus Integer     `snmp:"1.3.6.1.2.1.2.2.1.7"`
    OperStatus  Integer     `snmp:"1.3.6.1.2.1.2.2.1.8"`
    LastChange  TimeTicks   `snmp:"1.3.6.1.2.1.2.2.1.9"`
}

type InterfaceTable map[InterfaceIndex]*InterfaceEntry
