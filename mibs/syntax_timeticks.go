package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

type TimeTicks snmp.TimeTicks32

func (value TimeTicks) String() string {
	return fmt.Sprintf("%v", time.Duration(value)*10*time.Millisecond)
}

type TimeTicksSyntax struct{}

func (syntax TimeTicksSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax TimeTicksSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
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
