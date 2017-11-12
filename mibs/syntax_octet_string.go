package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"strings"
)

type OctetString []byte

func (value OctetString) String() string {
	var hex = make([]string, len(value))

	for i, b := range value {
		hex[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(hex, " ")
}

type OctetStringSyntax struct{}

func (syntax OctetStringSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax OctetStringSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		return OctetString(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
