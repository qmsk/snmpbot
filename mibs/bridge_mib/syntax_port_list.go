package bridge_mib

import (
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
)

type PortList []uint8

func (value PortList) List() []uint {
	var ports []uint

	for byteOffset, octet := range value {
		var bitOffset uint
		for bitOffset = 0; bitOffset < 8; bitOffset++ {
			var port = uint(byteOffset)*8 + bitOffset + 1
			var bit = octet&(1<<(8-bitOffset-1)) != 0

			if bit {
				ports = append(ports, port)
			}
		}
	}

	return ports
}

func (value PortList) Map() map[uint]bool {
	var ports = make(map[uint]bool)

	for byteOffset, octet := range value {
		var bitOffset uint
		for bitOffset = 0; bitOffset < 8; bitOffset++ {
			var port = uint(byteOffset)*8 + bitOffset + 1
			var bit = octet&(1<<(8-bitOffset-1)) != 0

			ports[port] = bit
		}
	}

	return ports
}

func (value PortList) String() string {
	return fmt.Sprintf("%v", value.List())
}

type PortListSyntax struct{}

func (syntax PortListSyntax) UnpackIndex(index []int) (mibs.Value, []int, error) {
	return nil, index, mibs.SyntaxIndexError{syntax, index}
}

func (syntax PortListSyntax) Unpack(varBind snmp.VarBind) (mibs.Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		return PortList(value), nil
	default:
		return nil, mibs.SyntaxError{syntax, value}
	}
}
