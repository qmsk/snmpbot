package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"strings"
)

type PhysAddress []byte

func (physAddress PhysAddress) String() string {
	var parts = make([]string, len(physAddress))

	for i, octet := range physAddress {
		parts[i] = fmt.Sprintf("%02x", octet)
	}

	return strings.Join(parts, ":")
}

func (value PhysAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(value.String())
}

type PhysAddressSyntax struct{}

func (syntax PhysAddressSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func (syntax PhysAddressSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		return PhysAddress(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
