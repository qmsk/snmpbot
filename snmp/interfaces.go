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

type InterfaceIndex Integer

type Interface struct {
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

func (self *Interface) lookup (oid OID) Type {
    if SNMP_interfaces_ifIndex.Match(oid) {
        return &self.Index
    } else if SNMP_interfaces_ifDescr.Match(oid) {
        return &self.Descr
    } else if SNMP_interfaces_ifType.Match(oid) {
        return &self.Type
    } else if SNMP_interfaces_ifMtu.Match(oid) {
        return &self.Mtu
    } else if SNMP_interfaces_ifSpeed.Match(oid) {
        return &self.Speed
    } else if SNMP_interfaces_ifPhysAddress.Match(oid) {
        return &self.PhysAddress
    } else if SNMP_interfaces_ifAdminStatus.Match(oid) {
        return &self.AdminStatus
    } else if SNMP_interfaces_ifOperStatus.Match(oid) {
        return &self.OperStatus
    } else if SNMP_interfaces_ifLastChange.Match(oid) {
        return &self.LastChange
    } else {
        return nil
    }
}

func (self *Interface) set (oid OID, snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    if field := self.lookup(oid); field == nil {
        return fmt.Errorf("OID not found")
    } else if !field.match(snmpType) {
        return fmt.Errorf("SNMP type mismatch")
    } else {
        log.Printf("snmp:Interface.set: %v %v\n", oid, field)

        field.set(snmpValue)

        return nil
    }
}

func (self *Client) Interfaces() (map[int]*Interface, error) {
    interfaces := make(map[int]*Interface)

    err := self.snmp.Walk(SNMP_interfaces_ifTable.String(), func (pdu gosnmp.SnmpPDU) error {
        item := parseOID(pdu.Name)

        oid, index := item.Index()


        iface, ifaceExists := interfaces[index]
        if !ifaceExists {
            iface = new(Interface)
            interfaces[index] = iface
        }

        if err := iface.set(oid, pdu.Type, pdu.Value); err != nil {
            log.Printf("snmp:Client.Interfaces: %v %v: %v\n", oid, pdu.Type, err)
        }

        return nil
    })

    return interfaces, err
}
