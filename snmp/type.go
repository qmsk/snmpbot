package snmp

import (
    "encoding/json"
    "fmt"
    "github.com/soniah/gosnmp"
    "time"
    wapsnmp "github.com/cdevr/WapSNMP"
)

/* Types */
type Value interface {
    // Set value from an SNMP object retrieved as an SNMP VarBind
    setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error
}

type Index interface {
    // Set value from a table-entry OID sub-identifier index
    // See RFC1442#7.7 SNMPv2 SMI, Mapping of the INDEX clause
    setIndex(oid OID) error

    // String representation
    String() string
}

type Syntax interface {
    // Parse a wapsnmp.BER value into a higher-level representation per the object syntax
    parseValue(snmpValue interface{}) (interface{}, error)
}

/* Errors */
type TypeError struct {
    Value           Value
    SnmpType        gosnmp.Asn1BER
}
func (self TypeError) Error() string {
    return fmt.Sprintf("Invalid SNMP type for %T: %v", self.Value, self.SnmpType)
}

type ValueError struct {
    Value           Value
    SnmpValue       interface{}
}
func (self ValueError) Error() string {
    return fmt.Sprintf("Invalid SNMP value for %T: %v", self.Value, self.SnmpValue)
}

type SyntaxError struct {
    Syntax          Syntax
    SnmpValue       interface{}
}
func (self SyntaxError) Error() string {
    return fmt.Sprintf("Invalid value for Syntax %T: %#v", self.Syntax, self.SnmpValue)
}

/* Integer */
type Integer int

func (self Integer) String() string {
    return fmt.Sprintf("%v", int(self))
}

func (self Integer) MarshalJSON() ([]byte, error) {
    return json.Marshal(int(self))
}

func (self *Integer) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
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

func (self Integer) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case int64:
        return Integer(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var IntegerSyntax Integer

/* String */
type String string

func (self String) String() string {
    return fmt.Sprintf("%s", string(self))
}

func (self String) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(self))
}

func (self *String) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.OctetString:
        value := snmpValue.([]byte)

        *self = String(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

func (self String) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case string:
        return String(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var StringSyntax String

/* Binary */
type Binary []byte

func (self Binary) String() string {
    return fmt.Sprintf("%x", []byte(self))
}

func (self Binary) MarshalJSON() ([]byte, error) {
    return json.Marshal([]byte(self))
}

func (self *Binary) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
    switch snmpType {
    case gosnmp.OctetString:
        value := snmpValue.([]byte)

        *self = Binary(value)
    default:
        return TypeError{self, snmpType}
    }

    return nil
}

/* ObjectID */
func (self OID) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case wapsnmp.Oid:
        return OID(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var OIDSyntax OID

/* Counter */
type Counter uint

func (self Counter) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Counter) MarshalJSON() ([]byte, error) {
    return json.Marshal(uint(self))
}

func (self *Counter) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
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

func (self *Gauge) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
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

func (self *TimeTicks) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
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

func (self TimeTicks) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case time.Duration:
        return TimeTicks(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var TimeTicksSyntax TimeTicks

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

func (self *MacAddress) setValue(snmpType gosnmp.Asn1BER, snmpValue interface{}) error {
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
