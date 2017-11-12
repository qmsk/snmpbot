package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type MACAddress [6]byte

func (value MACAddress) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		value[0],
		value[1],
		value[2],
		value[3],
		value[4],
		value[5],
	)
}

func (value MACAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(value.String())
}

type MACAddressSyntax struct{}

func (syntax MACAddressSyntax) UnpackIndex(index []int) (Value, []int, error) {
	if len(index) < 6 {
		return nil, index, SyntaxIndexError{syntax, index}
	}

	var value MACAddress

	for i := 0; i < 6; i++ {
		if index[i] < 0 || index[i] >= 256 {
			return nil, index, SyntaxIndexError{syntax, index[0:6]}
		}

		value[i] = byte(index[i])
	}

	return value, index[6:], nil
}

func (syntax MACAddressSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		var macAddress MACAddress

		if len(value) != 6 {
			return nil, SyntaxError{syntax, value}
		} else {
			copy(macAddress[:], value[0:6])
		}
		return macAddress, nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
