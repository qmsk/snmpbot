package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type IDKey string

type ID struct {
	MIB  *MIB
	Name string
	OID  snmp.OID
}

func (id ID) Key() IDKey {
	return IDKey(id.OID.String()) // TODO: perf?
}

func (id ID) String() string {
	if id.MIB == nil {
		return id.OID.String()
	} else if id.Name == "" {
		return id.MIB.FormatOID(id.OID)
	} else {
		return fmt.Sprintf("%s::%s", id.MIB.Name, id.Name)
	}
}

func (id ID) MakeID(name string, ids ...int) ID {
	return ID{id.MIB, name, id.OID.Extend(ids...)}
}

func (id ID) FormatOID(oid snmp.OID) string {
	if index := id.OID.Index(oid); index == nil {
		return oid.String()
	} else if len(index) == 0 {
		return id.String()
	} else {
		return fmt.Sprintf("%s::%s%s", id.MIB.Name, id.Name, snmp.OID(index).String())
	}
}
