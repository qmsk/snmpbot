package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Integer int

func (value Integer) String() string {
	return fmt.Sprintf("%v", int(value))
}

func (value Integer) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(value))
}

type IntegerSyntax struct{}

func (syntax IntegerSyntax) UnpackIndex(index []int) (Value, []int, error) {
	if len(index) < 1 {
		return nil, index, SyntaxIndexError{syntax, index}
	}

	return Integer(index[0]), index[1:], nil
}

func (syntax IntegerSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
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
