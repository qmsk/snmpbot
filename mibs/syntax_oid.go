package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type OID snmp.OID

func (value OID) String() string {
	return fmt.Sprintf("%v", snmp.OID(value))
}

func (value OID) MarshalJSON() ([]byte, error) {
	return json.Marshal(value.String())
}

type ObjectIdentifierSyntax struct{}

func (syntax ObjectIdentifierSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax ObjectIdentifierSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
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
