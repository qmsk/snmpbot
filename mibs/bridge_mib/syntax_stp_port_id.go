package bridge_mib

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
)

type PortID struct {
	Priority uint
	Index    uint
}

func (value PortID) String() string {
	return fmt.Sprintf("%d.%d", value.Priority, value.Index)
}

func (value PortID) MarshalJSON() ([]byte, error) {
	return json.Marshal(value.String())
}

type PortIDSyntax struct{}

func (syntax PortIDSyntax) UnpackIndex(index []int) (mibs.Value, []int, error) {
	// TODO
	return nil, index, mibs.SyntaxIndexError{syntax, index}
}

func (syntax PortIDSyntax) Unpack(varBind snmp.VarBind) (mibs.Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []byte:
		var portID PortID

		if len(value) != 2 {
			return nil, mibs.SyntaxError{syntax, value}
		} else {
			var uintValue uint16 = uint16(value[0])<<8 + uint16(value[1])

			portID.Priority = uint((uintValue & 0xf000) >> 8) // effectively * 16
			portID.Index = uint(uintValue & 0x0fff)
		}
		return portID, nil
	default:
		return nil, mibs.SyntaxError{syntax, value}
	}
}

func init() {
	// XXX: This is made up for BRIDGE-MIB::dot1dStpPortDesignatedPort
	mibs.RegisterSyntax("BRIDGE-MIB::PortId", PortIDSyntax{})
}
