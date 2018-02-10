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
	} else if snmpValue, err := varBind.Value(); err != nil {
		return nil, err
	} else if object.Syntax == nil {
		return snmpValue, nil
	} else {
		// TODO: change interface to Unpack(interface{})?
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

func (object *Object) FormatIndex(oid snmp.OID) (string, error) {
	if index := object.OID.Index(oid); index == nil {
		return oid.String(), nil
	} else if len(index) == 0 {
		return object.String(), nil
	} else if indexString, err := object.IndexSyntax.FormatIndex(index); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%s::%s%s", object.MIB.Name, object.Name, indexString), nil
	}
}

func (object *Object) Format(varBind snmp.VarBind) (string, Value, error) {
	if value, err := object.Unpack(varBind); err != nil {
		return "", nil, err
	} else if name, err := object.FormatIndex(varBind.OID()); err != nil {
		return "", value, err
	} else {
		return name, value, err
	}
}
