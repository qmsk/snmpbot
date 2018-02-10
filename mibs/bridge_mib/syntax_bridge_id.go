package bridge_mib

import (
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
)

type BridgeID struct {
	Priority uint
	mibs.MACAddress
}

func (value BridgeID) String() string {
	return fmt.Sprintf("%d@%v", value.Priority, value.MACAddress)
}

type BridgeIDSyntax struct{}

func (syntax BridgeIDSyntax) UnpackIndex(index []int) (mibs.Value, []int, error) {
	// TODO
	return nil, index, mibs.SyntaxIndexError{syntax, index}
}

func (syntax BridgeIDSyntax) Unpack(varBind snmp.VarBind) (mibs.Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		var bridgeID BridgeID

		if len(value) != 8 {
			return nil, mibs.SyntaxError{syntax, value}
		} else {
			bridgeID.Priority = uint(value[0])<<8 + uint(value[1])
			copy(bridgeID.MACAddress[:], value[2:8])
		}
		return bridgeID, nil
	default:
		return nil, mibs.SyntaxError{syntax, value}
	}
}

func init() {
	mibs.RegisterSyntax("BRIDGE-MIB::BridgeId", BridgeIDSyntax{})
}
