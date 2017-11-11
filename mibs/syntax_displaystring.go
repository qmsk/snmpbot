package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
)

type DisplayString string

type DisplayStringSyntax struct{}

func (syntax DisplayStringSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax DisplayStringSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
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
