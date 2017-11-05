package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type DisplayString string

func (value DisplayString) String() string {
	return fmt.Sprintf("%v", string(value))
}

func (syntax DisplayString) Unpack(varBind snmp.VarBind) (Value, error) {
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
