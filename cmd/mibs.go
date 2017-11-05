package cmd

import (
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
	_ "github.com/qmsk/snmpbot/mibs/snmpv2"
	"github.com/qmsk/snmpbot/snmp"
	"log"
)

func ParseOID(arg string) (snmp.OID, error) {
	return mibs.ParseOID(arg)
}

func (options Options) FormatOID(oid snmp.OID) string {
	return mibs.FormatOID(oid)
}

func (options Options) PrintVarBind(varBind snmp.VarBind) {
	if object := mibs.LookupObject(varBind.OID()); object != nil {
		options.PrintObject(object, varBind)
	} else if value, err := varBind.Value(); err != nil {
		log.Printf("VarBind[%v]: %v", varBind.OID(), err)
	} else {
		fmt.Printf("%v = <%T> %v\n", options.FormatOID(varBind.OID()), value, value)
	}
}

func (options Options) PrintObject(object *mibs.Object, varBind snmp.VarBind) {
	if name, value, err := object.Format(varBind); err != nil {
		fmt.Printf("%v = <%T> % !%v\n", name, value, value, err)
	} else {
		fmt.Printf("%v = %v\n", name, value)
	}
}
