package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type OID snmp.OID

func (value OID) String() string {
	return fmt.Sprintf("%v", snmp.OID(value))
}

func (syntax OID) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []int:
		return OID(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
