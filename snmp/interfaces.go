package snmp

import (
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
)

type Interface struct {
    IfIndex     int
    IfDescr     string
}

func (self *Interface) set (oid OID, value interface{}) {

    if indexValue, ok := SNMP_interfaces_ifIndex.MatchInteger(oid, value); ok {
        self.IfIndex = indexValue
    } else if descrValue, ok := SNMP_interfaces_ifDescr.MatchString(oid, value); ok {
        self.IfDescr = descrValue
    }
}

func (self *Client) Interfaces() (map[int]*Interface, error) {
    interfaces := make(map[int]*Interface)

    err := self.snmp.Walk(SNMP_interfaces_ifTable.String(), func (pdu gosnmp.SnmpPDU) error {
        item := parseOID(pdu.Name)

        oid, index := item.Index()

        log.Printf("snmp:Client.Interfaces: %+v\n", pdu)

        iface, ifaceExists := interfaces[index]
        if !ifaceExists {
            iface = new(Interface)
            interfaces[index] = iface
        }

        iface.set(oid, pdu.Value)

        return nil
    })

    return interfaces, err
}
