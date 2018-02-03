package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Value interface{}

type Syntax interface {
	UnpackIndex([]int) (Value, []int, error)
	Unpack(snmp.VarBind) (Value, error)
}

type SyntaxError struct {
	Syntax    Syntax
	SNMPValue interface{}
}

func (err SyntaxError) Error() string {
	return fmt.Sprintf("Invalid value for Syntax %T: <%T> %#v", err.Syntax, err.SNMPValue, err.SNMPValue)
}

type SyntaxIndexError struct {
	Syntax Syntax
	Index  []int
}

func (err SyntaxIndexError) Error() string {
	return fmt.Sprintf("Invalid index for Syntax %T: %#v", err.Syntax, err.Index)
}
