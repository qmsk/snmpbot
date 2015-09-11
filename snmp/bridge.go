package snmp

// SNMP BRIDGE-MIB implementation

var (
    BridgeMIB       = registerMIB("BRIDGE-MIB", 1,3,6,1,2,1,17)
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
