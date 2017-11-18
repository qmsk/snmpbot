package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Object struct {
	ID

	IndexSyntax
	Syntax
	NotAccessible bool
}

func (object *Object) Unpack(varBind snmp.VarBind) (Value, error) {
	if err := varBind.ErrorValue(); err != nil {
		return nil, err
	} else {
		return object.Syntax.Unpack(varBind)
	}
}

func (object *Object) UnpackIndex(oid snmp.OID) (IndexValues, error) {
	if oidIndex := object.OID.Index(oid); oidIndex == nil {
		return nil, fmt.Errorf("Invalid OID for Object<%v>: %v", oid, object)
	} else {
		return object.IndexSyntax.UnpackIndex(oidIndex)
	}
}

func (object *Object) Format(varBind snmp.VarBind) (string, Value, error) {
	name := object.FormatOID(varBind.OID())
	value, err := object.Unpack(varBind)

	return name, value, err
}

func (object *Object) FormatIndex(oid snmp.OID) string {
	if object.IndexSyntax == nil {
		return object.ID.FormatOID(oid)
	}

	if index := object.OID.Index(oid); index == nil {
		return oid.String()
	} else if len(index) == 0 {
		return object.String()
	} else if indexString, err := object.IndexSyntax.FormatIndex(index); err != nil {
		return fmt.Sprintf("%s::%s%s", object.MIB.Name, object.Name, snmp.OID(index).String())
	} else {
		return fmt.Sprintf("%s::%s%s", object.MIB.Name, object.Name, indexString)
	}
}
