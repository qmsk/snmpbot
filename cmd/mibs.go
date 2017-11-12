package cmd

import (
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
	_ "github.com/qmsk/snmpbot/mibs/bridge_mib"
	_ "github.com/qmsk/snmpbot/mibs/if_mib"
	_ "github.com/qmsk/snmpbot/mibs/snmpv2_mib"
	"github.com/qmsk/snmpbot/snmp"
	"log"
)

func ParseOID(arg string) (snmp.OID, error) {
	return mibs.ParseOID(arg)
}

func (options Options) ResolveID(name string) (mibs.ID, error) {
	return mibs.Resolve(name)
}

func (options Options) ResolveIDs(names []string) ([]mibs.ID, error) {
	var ids = make([]mibs.ID, len(names))

	for i, name := range names {
		if id, err := options.ResolveID(name); err != nil {
			return nil, fmt.Errorf("Invalid ID %v: %v", name, err)
		} else {
			ids[i] = id
		}
	}

	return ids, nil
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
	name := object.FormatIndex(varBind.OID())
	value, err := object.Unpack(varBind)

	if err != nil {
		log.Printf("VarBind[%v](%v): %v", varBind.OID(), object, err)
	} else {
		fmt.Printf("%v = %v\n", name, value)
	}
}
