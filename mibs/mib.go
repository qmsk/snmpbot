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
		tables:  make(map[*ID]*Table),
	}
}

type MIB struct {
	OID  snmp.OID
	Name string

	lookup  map[string]*ID
	resolve map[string]*ID
	objects map[*ID]*Object
	tables  map[*ID]*Table
}

func (mib *MIB) String() string {
	return mib.OID.String()
}

func (mib *MIB) MakeID(name string, ids ...int) ID {
	return ID{mib, name, mib.OID.Extend(ids...)}
}

func (mib *MIB) registerLookup(id *ID) {
	mib.lookup[id.OID.String()] = id
}

func (mib *MIB) registerResolve(id *ID) {
	mib.resolve[id.Name] = id
}

func (mib *MIB) RegisterObject(id ID, object Object) *Object {
	object.ID = &id

	mib.registerLookup(&id)
	mib.registerResolve(&id)
	mib.objects[&id] = &object

	return &object
}

func (mib *MIB) RegisterTable(id ID, table Table) *Table {
	table.ID = &id

	mib.registerResolve(&id)
	mib.tables[&id] = &table

	return &table
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

func (mib *MIB) LookupObject(oid snmp.OID) *Object {
	if id := mib.Lookup(oid); id == nil {
		return nil
	} else if object, ok := mib.objects[id]; !ok {
		return nil
	} else {
		return object
	}
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
