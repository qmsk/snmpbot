package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Enum struct {
	Value int
	Name  string
}

func (enum Enum) String() string {
	if enum.Name != "" {
		return enum.Name
	} else {
		return fmt.Sprintf("%d", enum.Value)
	}
}

func (enum Enum) MarshalJSON() ([]byte, error) {
	if enum.Name != "" {
		return json.Marshal(enum.Name)
	} else {
		return json.Marshal(enum.Value)
	}
}

type EnumSyntax []Enum

func (syntax EnumSyntax) lookup(value int) Enum {
	for _, enum := range syntax {
		if enum.Value == value {
			return enum
		}
	}

	return Enum{Value: value}
}

func (syntax EnumSyntax) Unpack(varBind snmp.VarBind) (Value, error) {
	snmpValue, err := varBind.Value()
	if err != nil {
		return nil, err
	}
	switch value := snmpValue.(type) {
	case int64:
		return syntax.lookup(int(value)), nil
	default:
		return nil, SyntaxError{syntax, value}
	}
}
