package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
)

type Object struct {
	*ID

	Syntax Syntax
}

func (object *Object) Unpack(varBind snmp.VarBind) (interface{}, error) {
	return object.Syntax.Unpack(varBind)
}

func (object *Object) Format(varBind snmp.VarBind) (string, interface{}, error) {
	name := object.FormatOID(varBind.OID())
	value, err := object.Unpack(varBind)

	return name, value, err
}
