package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

type Syntax interface {
	String() string
	Unpack(snmp.VarBind) (Syntax, error)
}

type SyntaxError struct {
	Syntax    Syntax
	SNMPValue interface{}
}

func (err SyntaxError) Error() string {
	return fmt.Sprintf("Invalid value for Syntax %T: <%T> %#v", err.Syntax, err.SNMPValue, err.SNMPValue)
}

type DisplayString string

func (value DisplayString) String() string {
	return string(value)
}

func (syntax DisplayString) Unpack(varBind snmp.VarBind) (Syntax, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		return DisplayString(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

type OID snmp.OID

func (value OID) String() string {
	return snmp.OID(value).String()
}

func (syntax OID) Unpack(varBind snmp.VarBind) (Syntax, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []int:
		return OID(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

type TimeTicks snmp.TimeTicks32

func (value TimeTicks) String() string {
	return fmt.Sprintf("%v", time.Duration(value)*10*time.Millisecond)
}

func (syntax TimeTicks) Unpack(varBind snmp.VarBind) (Syntax, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case snmp.TimeTicks32:
		return TimeTicks(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

type Integer int

func (value Integer) String() string {
	return fmt.Sprintf("%v", int(value))
}

func (syntax Integer) Unpack(varBind snmp.VarBind) (Syntax, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case int64:
		return Integer(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

var DisplayStringSyntax Syntax = DisplayString("")
var ObjectIdentifierSyntax Syntax = OID{}
var TimeTicksSyntax Syntax = TimeTicks(0)
var IntegerSyntax Syntax = Integer(0)
