package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Syntax interface {
	String() string
}

type DisplayString string

func (value DisplayString) String() string {
	return string(value)
}

type OID snmp.OID

func (value OID) String() string {
	return snmp.OID(value).String()
}

type TimeTicks snmp.TimeTicks32

func (value TimeTicks) String() string {
	return fmt.Sprintf("%v", snmp.TimeTicks32(value))
}

type Integer int

func (value Integer) String() string {
	return fmt.Sprintf("%v", int(value))
}

var DisplayStringSyntax Syntax = DisplayString("")
var ObjectIdentifierSyntax Syntax = OID{}
var TimeTicksSyntax Syntax = TimeTicks(0)
var IntegerSyntax Syntax = Integer(0)
