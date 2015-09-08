package snmp

import (
    "encoding/json"
    "fmt"
    "github.com/soniah/gosnmp"
    "time"
)

/* Types */
type Type interface {
    // Set value from an SNMP object retrieved as an SNMP VarBind
    set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error
}

type IndexType interface {
    // Set value from a table-entry OID sub-identifier index
    // See RFC1442#7.7 SNMPv2 SMI, Mapping of the INDEX clause
    setIndex(oid OID) error
}

type TypeError struct {
    Type            Type
    SnmpType        gosnmp.Asn1BER
}
func (self TypeError) Error() string {
    return fmt.Sprintf("Invalid SNMP type for %T: %v", self.Type, self.SnmpType)
}

type ValueError struct {
    Type            Type
    SnmpValue       interface{}
}
func (self ValueError) Error() string {
    return fmt.Sprintf("Invalid SNMP value for %T: %v", self.Type, self.SnmpValue)
}

/* Integer */
type Integer int

func (self Integer) String() string {
    return fmt.Sprintf("%v", int(self))
}

func (self Integer) MarshalJSON() ([]byte, error) {
    return json.Marshal(int(self))
}

func (self *Integer) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.Integer:
        value := snmpValue.(int)

        *self = Integer(value)

    default:
        return TypeError{self, snmpType}
    }

    return nil
}

func (self *Integer) setIndex(oid OID) error {
    *self = Integer(oid[0])

    return nil
}

/* String */
type String string

func (self String) String() string {
    return fmt.Sprintf("%s", string(self))
}

func (self String) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(self))
}

func (self *String) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.OctetString:
        value := snmpValue.([]byte)

        *self = String(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* Binary */
type Binary []byte

func (self Binary) String() string {
    return fmt.Sprintf("%x", []byte(self))
}

func (self Binary) MarshalJSON() ([]byte, error) {
    return json.Marshal([]byte(self))
}

func (self *Binary) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.OctetString:
        value := snmpValue.([]byte)

        *self = Binary(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* Counter */
type Counter uint

func (self Counter) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Counter) MarshalJSON() ([]byte, error) {
    return json.Marshal(uint(self))
}

func (self *Counter) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.Counter32:
        value := snmpValue.(uint)

        *self = Counter(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* Gauge */
type Gauge uint

func (self Gauge) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Gauge) MarshalJSON() ([]byte, error) {
    return json.Marshal(uint(self))
}

func (self *Gauge) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.Gauge32:
        value := snmpValue.(uint)

        *self = Gauge(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* TimeTicks */
type TimeTicks time.Duration

func (self TimeTicks) String() string {
    return fmt.Sprintf("%v", time.Duration(self))
}

func (self TimeTicks) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(self))
}

func (self *TimeTicks) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.TimeTicks:
        value := snmpValue.(int)

        // convert from 100ths of a second
        duration := time.Duration(value * 10) * time.Millisecond

        *self = TimeTicks(duration)

    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* MacAddress */
type MacAddress [6]byte

func (self MacAddress) String() string {
    return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
        self[0],
        self[1],
        self[2],
        self[3],
        self[4],
        self[5],
    )
}

func (self MacAddress) MarshalJSON() ([]byte, error) {
    return json.Marshal(self.String())
}

func (self *MacAddress) set(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.OctetString:
        value := snmpValue.([]byte)

        if len(value) != 6 {
            return ValueError{self, snmpValue}
        }

        for i := 0; i < 6; i++ {
            self[i] = byte(value[i])
        }

    default:
        return TypeError{self, snmpType}
    }

    return nil
}

func (self *MacAddress) setIndex(oid OID) error {
    if len(oid) != 6 {
        return fmt.Errorf("Invalid sub-OID for %T index: %v", self, oid)
    }

    for i := 0; i < 6; i++ {
        self[i] = byte(oid[i])
    }

    return nil
}
