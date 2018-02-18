package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Counter uint

func (value Counter) String() string {
	return fmt.Sprintf("%v", uint(value))
}

type CounterSyntax struct{}

func (syntax CounterSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case snmp.Counter32:
		return Counter(value), nil
	case snmp.Counter64:
		return Counter(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

func (syntax CounterSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func init() {
	RegisterSyntax("Counter", CounterSyntax{})
	RegisterSyntax("Counter32", CounterSyntax{})
	RegisterSyntax("Counter64", CounterSyntax{})
}
