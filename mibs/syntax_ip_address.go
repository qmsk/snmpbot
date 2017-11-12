package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
	"net"
)

type IPAddress net.IP

func (value IPAddress) String() string {
	return net.IP(value).String()
}

type IPAddressSyntax struct{}

func (syntax IPAddressSyntax) UnpackIndex(index []int) (Value, []int, error) {
	if len(index) < 4 {
		return nil, index, SyntaxIndexError{syntax, index}
	}

	var value = make(IPAddress, 4)

	for i := 0; i < 4; i++ {
		if index[i] < 0 || index[i] >= 256 {
			return nil, index, SyntaxIndexError{syntax, index[0:4]}
		}

		value[i] = byte(index[i])
	}

	return value, index[4:], nil
}

func (syntax IPAddressSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case snmp.IPAddress:
		return IPAddress(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
