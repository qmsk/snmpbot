package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Unsigned uint

func (value Unsigned) String() string {
	return fmt.Sprintf("%v", uint(value))
}

type UnsignedSyntax struct{}

func (syntax UnsignedSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case int:
    if value <= 0 {
      return nil, SyntaxError{syntax, value}
    }
		return Unsigned(value), nil
	case snmp.Gauge32:
		return Unsigned(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

func (syntax UnsignedSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}
