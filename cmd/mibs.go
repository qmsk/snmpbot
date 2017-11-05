package cmd

import (
	"github.com/qmsk/snmpbot/mibs"
	_ "github.com/qmsk/snmpbot/mibs/snmpv2"
	"github.com/qmsk/snmpbot/snmp"
)

func ParseOID(arg string) (snmp.OID, error) {
	return mibs.ParseOID(arg)
}

func FormatOID(oid snmp.OID) string {
	return mibs.FormatOID(oid)
}
