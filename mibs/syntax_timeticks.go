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

func (syntax TimeTicks) Unpack(varBind snmp.VarBind) (Value, error) {
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
