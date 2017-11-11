package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
)

type OctetString []byte

type OctetStringSyntax struct{}

func (syntax OctetStringSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax OctetStringSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		return OctetString(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
