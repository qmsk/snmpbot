package snmp

// SNMP IF-MIB implementation

var (
    InterfacesMIB       = MIB{OID{1,3,6,1,2,1,2}}
)

type InterfaceIndex struct {
    Index       Integer
}

func (self *InterfaceIndex) setIndex (oid OID) error {
    return self.Index.setIndex(oid)
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

func (self *Client) Interfaces() (table InterfaceTable, err error) {
    table = make(InterfaceTable)

    err = self.GetTable(&table)

    return
}
