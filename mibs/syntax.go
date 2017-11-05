package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Value interface{}

type Syntax interface {
	Unpack(snmp.VarBind) (Value, error)
}

type IndexSyntax interface {
	UnpackIndex([]int) (Value, []int, error)
}

type SyntaxError struct {
	Syntax    Syntax
	SNMPValue interface{}
}

func (err SyntaxError) Error() string {
	return fmt.Sprintf("Invalid value for Syntax %T: <%T> %#v", err.Syntax, err.SNMPValue, err.SNMPValue)
}

type IndexError struct {
	Syntax IndexSyntax
	Index  []int
}

func (err IndexError) Error() string {
	return fmt.Sprintf("Invalid value for IndexSyntax %T: %#v", err.Syntax, err.Index)
}

var DisplayStringSyntax Syntax = DisplayString("")
var ObjectIdentifierSyntax Syntax = OID{}
var PhysAddressSyntax Syntax = PhysAddress{}
var GaugeSyntax Syntax = Gauge(0)
var TimeTicksSyntax Syntax = TimeTicks(0)
var IntegerSyntax Syntax = Integer(0)
var IntegerIndexSyntax IndexSyntax = Integer(0)
