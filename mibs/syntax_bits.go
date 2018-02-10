package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Bit struct {
	Bit  uint
	Name string
}

func (bit Bit) String() string {
	if bit.Name != "" {
		return bit.Name
	} else {
		return fmt.Sprintf("%d", 1<<bit.Bit)
	}
}

func (bit Bit) MarshalJSON() ([]byte, error) {
	if bit.Name != "" {
		return json.Marshal(bit.Name)
	} else {
		return json.Marshal(1 << bit.Bit)
	}
}

type BitsValue []Bit

type BitsSyntax []Bit

func (syntax BitsSyntax) values(value []uint8) BitsValue {
	var values = make(BitsValue, 0)

	for _, bit := range syntax {
		var byteOffset = bit.Bit / uint(8)
		var bitOffset = bit.Bit % uint(8)

		if uint(len(value)) <= byteOffset {
			continue
		}

		var byteValue = value[byteOffset]

		if byteValue&(0x80>>bitOffset) != 0 {
			values = append(values, bit)
		}
	}

	return values
}

func (syntax BitsSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case []uint8:
		return syntax.values(value), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}

func (syntax BitsSyntax) UnpackIndex(index []int) (Value, []int, error) {
	// TODO
	return nil, index, SyntaxIndexError{syntax, index}
}

func init() {
	RegisterSyntax("BITS", BitsSyntax{})
}
