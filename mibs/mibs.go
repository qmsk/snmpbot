package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"strings"
)

var registry = Registry{
	lookup:  make(map[string]*MIB),
	resolve: make(map[string]*MIB),
}

func RegisterMIB(name string, oid snmp.OID) *MIB {
	var mib = newMIB(name, oid)

	registry.Register(mib)

	return mib
}

// Lookup human-readable object name with optional index
func ParseOID(name string) (snmp.OID, error) {
	var mib *MIB
	var id *ID
	var index []int

	if name == "" {
		return nil, nil
	}

	if parts := strings.SplitN(name, "::", 2); len(parts) == 1 && name[0] == '.' {
		// .X.Y.Z
	} else if resolveMIB := registry.Resolve(parts[0]); resolveMIB == nil {
		return nil, fmt.Errorf("MIB not found: %v", parts[0])
	} else {
		mib = resolveMIB

		if len(parts) == 1 {
			// MIB
			name = ""
		} else {
			// MIB::*
			name = parts[1]
		}
	}

	if mib == nil {

	} else if name == "" {

	} else if parts := strings.SplitN(name, ".", 2); len(parts) > 1 && name[0] == '.' {
		// MIB::.X.Y.Z
	} else if resolveID := mib.Resolve(parts[0]); resolveID == nil {
		return nil, fmt.Errorf("%v name not found: %v", mib.Name, parts[0])
	} else {
		id = resolveID

		if len(parts) == 1 {
			// MIB::name
			name = ""
		} else {
			// MIB::name.X
			name = "." + parts[1]
		}
	}

	if name == "" {

	} else if oid, err := snmp.ParseOID(name); err != nil {
		return nil, err
	} else {
		index = []int(oid)
	}

	if mib == nil {
		return snmp.OID(index), nil
	} else if id == nil {
		return mib.OID.Extend(index...), nil
	} else if index != nil {
		return id.OID.Extend(index...), nil
	} else {
		return id.OID, nil
	}
}

// Lookup machine-readable object ID with optional index
func Lookup(oid snmp.OID) (*MIB, *ID) {
	if mib := registry.Lookup(oid); mib == nil {
		return nil, nil
	} else {
		return mib, mib.Lookup(oid)
	}
}

// Lookup machine-readable object ID with optional index
func LookupObject(oid snmp.OID) *Object {
	if mib := registry.Lookup(oid); mib == nil {
		return nil
	} else {
		return mib.LookupObject(oid)
	}
}

func FormatOID(oid snmp.OID) string {
	mib, id := Lookup(oid)

	if mib == nil {
		return oid.String()
	} else if id == nil {
		return mib.FormatOID(oid)
	} else {
		return id.FormatOID(oid)
	}
}
