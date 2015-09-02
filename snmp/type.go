package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "time"
)

/* Types */
type Type interface {
    match(snmpType gosnmp.Asn1BER) bool
    set(snmpValue interface{})
}

/* Integer */
type Integer int

func (self Integer) String() string {
    return fmt.Sprintf("%v", int(self))
}

func (self Integer) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.Integer
}

func (self *Integer) set(snmpValue interface{}) {
    value := snmpValue.(int)

    *self = Integer(value)
}

/* String */
type String string

func (self String) String() string {
    return fmt.Sprintf("%s", string(self))
}

func (self String) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.OctetString
}

func (self *String) set(snmpValue interface{}) {
    value := snmpValue.([]byte)

    *self = String(value)
}

/* Binary */
type Binary []byte

func (self Binary) String() string {
    return fmt.Sprintf("%x", []byte(self))
}

func (self Binary) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.OctetString
}

func (self *Binary) set(snmpValue interface{}) {
    value := snmpValue.([]byte)

    *self = Binary(value)
}

/* Counter */
type Counter uint

func (self Counter) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Counter) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.Counter32
}

func (self *Counter) set(snmpValue interface{}) {
    value := snmpValue.(uint)

    *self = Counter(value)
}

/* Gauge */
type Gauge uint

func (self Gauge) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Gauge) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.Gauge32
}

func (self *Gauge) set(snmpValue interface{}) {
    value := snmpValue.(uint)

    *self = Gauge(value)
}

/* TimeTicks */
type TimeTicks time.Duration

func (self TimeTicks) String() string {
    return fmt.Sprintf("%v", time.Duration(self))
}

func (self TimeTicks) match(snmpType gosnmp.Asn1BER) bool {
    return snmpType == gosnmp.TimeTicks
}

func (self *TimeTicks) set(snmpValue interface{}) {
    value := snmpValue.(int)

    // convert from 100ths of a second
    duration := time.Duration(value * 10) * time.Millisecond

    *self = TimeTicks(duration)
}
