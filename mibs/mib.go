package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

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

func (mib *MIB) Lookup(oid snmp.OID) *ID {
	var lookup = ""

	for _, id := range oid {
		lookup += fmt.Sprintf(".%d", id)

		if id := mib.lookup[lookup]; id != nil {
			return id
		}
	}

	return nil
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
