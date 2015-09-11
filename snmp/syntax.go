package snmp

import (
    "encoding/json"
    "fmt"
    "net"
    "time"
    wapsnmp "github.com/cdevr/WapSNMP"
)

/* Types */
type IndexSyntax interface {
    // parse index length
    peekIndex(oid OID) int

    // Set value from a table-entry OID sub-identifier index
    // See RFC1442#7.7 SNMPv2 SMI, Mapping of the INDEX clause
    parseIndex(oid OID) (IndexSyntax, error)

    // String representation
    String() string
}

type Syntax interface {
    // Parse a wapsnmp.BER value into a higher-level representation per the object syntax
    parseValue(snmpValue interface{}) (interface{}, error)
}

/* Errors */
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

func (self Integer) peekIndex(oid OID) int {
    return 1
}

func (self Integer) parseIndex(oid OID) (IndexSyntax, error) {
    if len(oid) != 1 {
        return nil, fmt.Errorf("Invalid sub-OID for %T index: %v", self, oid)
    }

    return Integer(oid[0]), nil
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

func (self Binary) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case string:
        return Binary(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var BinarySyntax Binary


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

func (self Counter) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case wapsnmp.Counter:
        return Counter(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var CounterSyntax Counter

/* Gauge */
type Gauge uint

func (self Gauge) String() string {
    return fmt.Sprintf("%v", uint(self))
}

func (self Gauge) MarshalJSON() ([]byte, error) {
    return json.Marshal(uint(self))
}

func (self Gauge) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case wapsnmp.Gauge:
        return Gauge(value), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var GaugeSyntax Gauge

/* TimeTicks */
type TimeTicks time.Duration

func (self TimeTicks) String() string {
    return fmt.Sprintf("%v", time.Duration(self))
}

func (self TimeTicks) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(self))
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

/* IPAddress */
type IpAddress net.IP

func (self IpAddress) String() string {
    return net.IP(self).String()
}

func (self IpAddress) MarshalJSON() ([]byte, error) {
    return json.Marshal(self.String())
}

func (self IpAddress) peekIndex(oid OID) int {
    return 4
}

func (self IpAddress) parseIndex(oid OID) (IndexSyntax, error) {
    if len(oid) != 4 {
        return nil, fmt.Errorf("Invalid sub-OID for %T index: %v", self, oid)
    }

    value := make(IpAddress, 4)

    for i := 0; i < 4; i++ {
        value[i] = byte(oid[i])
    }

    return value, nil
}

func (self IpAddress) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case net.IP:
        if ip4 := value.To4(); ip4 == nil {
            return nil, SyntaxError{self, snmpValue}
        } else {
            return IpAddress(ip4), nil
        }
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var IpAddressSyntax IpAddress

/* PhysAddress */
type PhysAddress []byte

func (self PhysAddress) String() string {
    str := ""

    for _, octet := range self {
        if str != "" {
            str += ":"
        }

        str += fmt.Sprintf("%02x", octet)
    }

    return str
}

func (self PhysAddress) MarshalJSON() ([]byte, error) {
    return json.Marshal(self.String())
}

func (self PhysAddress) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case string:
        return PhysAddress(([]byte)(value)), nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var PhysAddressSyntax PhysAddress

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

func (self MacAddress) peekIndex(oid OID) int {
    return 6
}

func (self MacAddress) parseIndex(oid OID) (IndexSyntax, error) {
    if len(oid) != 6 {
        return nil, fmt.Errorf("Invalid sub-OID for %T index: %v", self, oid)
    }

    value := MacAddress{}

    for i := 0; i < 6; i++ {
        value[i] = byte(oid[i])
    }

    return value, nil
}

func (self MacAddress) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case string:
        if len(value) != 6 {
            return nil, SyntaxError{self, snmpValue}
        }

        var retValue MacAddress

        for i := 0; i < 6; i++ {
            retValue[i] = byte(value[i])
        }

        return retValue, nil
    default:
        return nil, SyntaxError{self, snmpValue}
    }
}

var MacAddressSyntax MacAddress

/* Enum */
type Enum struct{
    Value       int
    Name        string
}

func (self Enum) String() string {
    if self.Name != "" {
        return self.Name
    } else {
        return fmt.Sprintf("%d", self.Value)
    }
}

func (self Enum) MarshalJSON() ([]byte, error) {
    if self.Name != "" {
        return json.Marshal(self.Name)
    } else {
        return json.Marshal(self.Value)
    }
}

type EnumSyntax []Enum

func (self EnumSyntax) parseValue(snmpValue interface{}) (interface{}, error) {
    switch value := snmpValue.(type) {
    case int64:
        enumValue := int(value)

        for _, enum := range self {
            if enum.Value == enumValue {
                return enum, nil
            }
        }

        return Enum{Value: enumValue}, nil

    default:
        return nil, SyntaxError{self, snmpValue}
    }
}
