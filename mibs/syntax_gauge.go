package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Gauge snmp.Gauge32

func (value Gauge) String() string {
	return fmt.Sprintf("%v", snmp.Gauge32(value))
}

func (syntax Gauge) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case snmp.Gauge32:
		return Gauge(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
