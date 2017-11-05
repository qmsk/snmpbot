package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type ID struct {
	MIB  *MIB
	Name string
	OID  snmp.OID
}

func (id *ID) String() string {
	return fmt.Sprintf("%s::%s", id.MIB.Name, id.Name)
}

func (id *ID) FormatOID(oid snmp.OID) string {
	if index := id.OID.Index(oid); index == nil {
		return oid.String()
	} else if len(index) == 0 {
		return id.String()
	} else {
		return fmt.Sprintf("%s::%s%s", id.MIB.Name, id.Name, snmp.OID(index).String())
	}
}
