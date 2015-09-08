package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "log"
)

var (
    SNMP_interfaces             = MIB{OID{1,3,6,1,2,1,2}}

    SNMP_interfaces_ifNumber    = SNMP_interfaces.define(1)
    SNMP_interfaces_ifTable     = SNMP_interfaces.define(2)
    SNMP_interfaces_ifEntry     = SNMP_interfaces_ifTable.define(1)
    SNMP_interfaces_ifIndex     = SNMP_interfaces_ifEntry.define(1)
    SNMP_interfaces_ifDescr     = SNMP_interfaces_ifEntry.define(2)
    SNMP_interfaces_ifType      = SNMP_interfaces_ifEntry.define(3)
    SNMP_interfaces_ifMtu       = SNMP_interfaces_ifEntry.define(4)
    SNMP_interfaces_ifSpeed     = SNMP_interfaces_ifEntry.define(5)
    SNMP_interfaces_ifPhysAddress   = SNMP_interfaces_ifEntry.define(6)
    SNMP_interfaces_ifAdminStatus   = SNMP_interfaces_ifEntry.define(7)
    SNMP_interfaces_ifOperStatus    = SNMP_interfaces_ifEntry.define(8)
    SNMP_interfaces_ifLastChange    = SNMP_interfaces_ifEntry.define(9)
)

type InterfaceIndex struct {
    Index       Integer
}

type InterfaceEntry struct {
    Index       Integer // SNMP_interfaces_ifIndex
    Descr       String  // SNMP_interfaces_ifDescr
    Type        Integer
    Mtu         Integer
    Speed       Gauge
    PhysAddress Binary
    AdminStatus Integer
    OperStatus  Integer
    LastChange  TimeTicks
}

type InterfaceTable map[InterfaceIndex]*InterfaceEntry

func (self *InterfaceEntry) field (id int) Type {
    switch id {
    case 1:
        return &self.Index
    case 2:
        return &self.Descr
    case 3:
        return &self.Type
    case 4:
        return &self.Mtu
    case 5:
        return &self.Speed
    case 6:
        return &self.PhysAddress
    case 7:
        return &self.AdminStatus
    case 8:
        return &self.OperStatus
    case 9:
        return &self.LastChange
    default:
        return nil
    }
}

func (self *InterfaceIndex) setIndex (oid OID) error {
    if err := self.Index.setIndex(oid); err != nil {
        return err
    }

    return nil
}

func (self InterfaceTable) set (oid OID, snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    selfMap := map[InterfaceIndex]*InterfaceEntry(self)

    var entryOid OID

    if entryOid = SNMP_interfaces_ifEntry.Index(oid); entryOid == nil {
        return fmt.Errorf("invalid")
    }

    var fieldId int = entryOid[0]
    var indexOid OID = entryOid[1:]

    // entry
    var index InterfaceIndex

    if err := index.setIndex(indexOid); err != nil {
        return err
    }

    entry, entryExists := selfMap[index]
    if !entryExists {
        entry = new(InterfaceEntry)
        selfMap[index] = entry
    }

    // field
    var field Type

    if field = entry.field(fieldId); field == nil {
        return fmt.Errorf("unknown")
    }

    // value
    if err := field.set(snmpType, snmpValue); err != nil {
        return err
    }

    return nil
}

func (self *Client) Interfaces() (InterfaceTable, error) {
    interfaces := make(InterfaceTable)

    err := self.snmp.Walk(SNMP_interfaces_ifTable.String(), func (pdu gosnmp.SnmpPDU) error {
        oid := parseOID(pdu.Name)

        if err := interfaces.set(oid, pdu.Type, pdu.Value); err != nil {
            log.Printf("snmp:Client.Interfaces: %v %v: %v\n", oid, pdu.Type, err)
        }

        return nil
    })

    return interfaces, err
}
