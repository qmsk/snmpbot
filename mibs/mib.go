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

type Registry struct {
	lookup  map[string]*MIB
	resolve map[string]*MIB
}

func (registry *Registry) Register(mib *MIB) {
	registry.lookup[mib.OID.String()] = mib
	registry.resolve[mib.Name] = mib
}

func (registry *Registry) Resolve(name string) *MIB {
	return registry.resolve[name]
}

func (registry *Registry) Lookup(oid snmp.OID) (*MIB, []int) {
	var lookup = ""

	for i, id := range oid {
		lookup += fmt.Sprintf(".%d", id)

		if mib := registry.lookup[lookup]; mib != nil {
			return mib, oid[i:]
		}
	}

	return nil, oid[:]
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
func Lookup(oid snmp.OID) (*MIB, *ID, []int) {
	if mib, index := registry.Lookup(oid); mib == nil {
		return nil, nil, index
	} else if id, index := mib.Lookup(oid); id == nil {
		return mib, nil, index
	} else {
		return mib, id, index
	}
}

func FormatOID(oid snmp.OID) string {
	mib, id, _ := Lookup(oid)

	if mib == nil {
		return oid.String()
	} else if id == nil {
		return mib.FormatOID(oid)
	} else {
		return id.FormatOID(oid)
	}
}

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

func newMIB(name string, oid snmp.OID) *MIB {
	return &MIB{
		OID:     oid,
		Name:    name,
		lookup:  make(map[string]*ID),
		resolve: make(map[string]*ID),
		objects: make(map[*ID]*Object),
	}
}

type MIB struct {
	OID  snmp.OID
	Name string

	lookup  map[string]*ID
	resolve map[string]*ID
	objects map[*ID]*Object
}

func (mib *MIB) String() string {
	return mib.OID.String()
}

func (mib *MIB) Register(name string, oid ...int) *ID {
	var id = &ID{mib, name, mib.OID.Extend(oid...)}

	mib.lookup[id.OID.String()] = id
	mib.resolve[id.Name] = id

	return id
}

func (mib *MIB) RegisterObject(id *ID, object Object) *Object {
	object.ID = id

	mib.objects[id] = &object

	return &object
}

func (mib *MIB) Resolve(name string) *ID {
	return mib.resolve[name]
}

func (mib *MIB) Lookup(oid snmp.OID) (*ID, []int) {
	var lookup = ""

	for i, id := range oid {
		lookup += fmt.Sprintf(".%d", id)

		if id := mib.lookup[lookup]; id != nil {
			return id, oid[i:]
		}
	}

	return nil, oid[:]
}

func (mib *MIB) FormatOID(oid snmp.OID) string {
	if index := mib.OID.Index(oid); index == nil {
		return oid.String()
	} else if len(index) == 0 {
		return mib.Name
	} else {
		return fmt.Sprintf("%s::%s", mib.Name, snmp.OID(index).String())
	}
}
