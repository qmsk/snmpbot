package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

type TimeTicks time.Duration

func unpackTimeTicks(value int) TimeTicks {
	return TimeTicks(time.Duration(value) * 10 * time.Millisecond)
}

func (value TimeTicks) Seconds() float64 {
	return time.Duration(value).Seconds()
}

func (value TimeTicks) String() string {
	return fmt.Sprintf("%v", time.Duration(value))
}

func (value TimeTicks) MarshalJSON() ([]byte, error) {
	return json.Marshal(value.Seconds())
}

type TimeTicksSyntax struct{}

func (syntax TimeTicksSyntax) UnpackIndex(index []int) (Value, []int, error) {
	if len(index) < 1 || index[0] < 0 {
		return nil, index, SyntaxIndexError{syntax, index}
	}

	return unpackTimeTicks(index[0]), index[1:], nil
}

func (syntax TimeTicksSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case snmp.TimeTicks32:
		return unpackTimeTicks(int(value)), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
