package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"regexp"
)

var mibRegistry = makeRegistry()

func registerMIB(mib MIB) *MIB {
	mib.ID.MIB = &mib

	mibRegistry.registerName(ID{MIB: &mib, OID: mib.OID}, mib.Name)
	mibRegistry.registerOID(ID{MIB: &mib, OID: mib.OID})

	return &mib
}

func RegisterMIB(name string, oid ...int) *MIB {
	return registerMIB(makeMIB(ID{Name: name, OID: snmp.OID(oid)}))
}

func ResolveMIB(name string) (*MIB, error) {
	if id, ok := mibRegistry.getName(name); !ok {
		return nil, fmt.Errorf("MIB not found: %v", name)
	} else {
		return id.MIB, nil
	}
}

func WalkMIBs(f func(mib *MIB)) {
	mibRegistry.walk(func(id ID) {
		f(id.MIB)
	})
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
		id.Name = "" // fixup MIB.ID re-use of Name
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

func ResolveObject(name string) (*Object, error) {
	if id, err := Resolve(name); err != nil {
		return nil, err
	} else if id.MIB == nil {
		return nil, fmt.Errorf("No MIB for name: %v", name)
	} else if object := id.MIB.Object(id); object == nil {
		return nil, fmt.Errorf("Not an object: %v", name)
	} else {
		return object, nil
	}
}

func ResolveTable(name string) (*Table, error) {
	if id, err := Resolve(name); err != nil {
		return nil, err
	} else if id.MIB == nil {
		return nil, fmt.Errorf("No MIB for name: %v", name)
	} else if table := id.MIB.Table(id); table == nil {
		return nil, fmt.Errorf("Not a table: %v", name)
	} else {
		return table, nil
	}
}

// Lookup ID by OID
func LookupMIB(oid snmp.OID) *MIB {
	if id, ok := mibRegistry.getOID(oid); !ok {
		return nil
	} else {
		return id.MIB
	}
}

func Lookup(oid snmp.OID) ID {
	if id, ok := mibRegistry.getOID(oid); !ok {
		return ID{OID: oid}
	} else {
		return id
	}
}

func LookupObject(oid snmp.OID) *Object {
	if id := Lookup(oid); id.MIB == nil {
		return nil
	} else {
		return id.MIB.Object(id)
	}
}

func Walk(f func(i ID)) {
	mibRegistry.walk(func(id ID) {
		id.MIB.Walk(f)
	})
}

func WalkObjects(f func(object *Object)) {
	Walk(func(id ID) {
		if object := id.MIB.Object(id); object != nil {
			f(object)
		}
	})
}

func WalkTables(f func(table *Table)) {
	Walk(func(id ID) {
		if table := id.MIB.Table(id); table != nil {
			f(table)
		}
	})
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
