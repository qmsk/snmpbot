package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"regexp"
)

var mibRegistry = makeRegistry()

func RegisterMIB(name string, oid snmp.OID) *MIB {
	var mib = makeMIB(ID{Name: name, OID: oid})

	mib.ID.MIB = &mib

	mibRegistry.register(mib.ID)

	return &mib
}

func ResolveMIB(name string) (*MIB, error) {
	if id, ok := mibRegistry.getName(name); !ok {
		return nil, fmt.Errorf("MIB not found: %v", name)
	} else {
		return id.MIB, nil
	}
}

func LookupMIB(oid snmp.OID) *MIB {
	if id, ok := mibRegistry.getOID(oid); !ok {
		return nil
	} else {
		return id.MIB
	}
}

/* Resolve ID by human-readable name:
 		".1.3.6"
		"SNMPv2-MIB"
		"SNMPv2-MIB.1.0"
		"SNMPv2-MIB::sysDescr"
		"SNMPv2-MIB::sysDescr.0"
*/
var resolveRegexp = regexp.MustCompile("^([^.:]+?)?(?:::([^.]+?))?([.][0-9.]+)?$")

func Resolve(name string) (ID, error) {
	var id ID
	var nameMIB, nameID, nameOID string

	if matches := resolveRegexp.FindStringSubmatch(name); matches == nil {
		return id, fmt.Errorf("Invalid syntax: %v", name)
	} else {
		nameMIB = matches[1]
		nameID = matches[2]
		nameOID = matches[3]
	}

	if nameMIB == "" {

	} else if mib, err := ResolveMIB(nameMIB); err != nil {
		return id, err
	} else {
		id = mib.ID
	}

	if nameID == "" {

	} else if id.MIB == nil {
		return id, fmt.Errorf("Invalid name without MIB: %v", name)
	} else if resolveID, err := id.MIB.ResolveName(nameID); err != nil {
		return id, err
	} else {
		id = resolveID
	}

	if nameOID == "" {

	} else if oid, err := snmp.ParseOID(nameOID); err != nil {
		return id, err
	} else {
		if id.OID == nil {
			id.OID = oid
		} else {
			id.OID = id.OID.Extend(oid...)
		}
	}

	return id, nil
}

// Lookup ID by OID
func Lookup(oid snmp.OID) ID {
	if mib := LookupMIB(oid); mib == nil {
		return ID{OID: oid}
	} else {
		return mib.Lookup(oid)
	}
}

func LookupObject(oid snmp.OID) *Object {
	if id := Lookup(oid); id.MIB == nil {
		return nil
	} else {
		return id.MIB.GetObject(id)
	}
}

// Lookup human-readable object name with optional index
func ParseOID(name string) (snmp.OID, error) {
	if id, err := Resolve(name); err != nil {
		return nil, err
	} else {
		return id.OID, nil
	}
}

func FormatOID(oid snmp.OID) string {
	return Lookup(oid).FormatOID(oid)
}
