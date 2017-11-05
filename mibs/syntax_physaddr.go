package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"strings"
)

type PhysAddress []byte

func (addr PhysAddress) String() string {
	var parts = make([]string, len(addr))

	for i, octet := range addr {
		parts[i] = fmt.Sprintf("%02x", octet)
	}

	return strings.Join(parts, ":")
}

func (syntax PhysAddress) Unpack(varBind snmp.VarBind) (Value, error) {
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
