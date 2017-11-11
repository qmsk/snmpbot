package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

func makeMIB(id *ID) MIB {
	return MIB{
		ID: id,
		registry: makeRegistry(),
		objects: make(map[*ID]*Object),
		tables:  make(map[*ID]*Table),
	}
}

type MIB struct {
	*ID
	registry

	objects map[*ID]*Object
	tables  map[*ID]*Table
}

func (mib *MIB) String() string {
	return mib.OID.String()
}

func (mib *MIB) MakeID(name string, ids ...int) ID {
	return ID{mib, name, mib.OID.Extend(ids...)}
}

func (mib *MIB) RegisterObject(id ID, object Object) *Object {
	object.ID = &id

	mib.registry.registerOID(&id)
	mib.registry.registerName(&id)

	mib.objects[&id] = &object

	return &object
}

func (mib *MIB) RegisterTable(id ID, table Table) *Table {
	table.ID = &id

	mib.registry.registerName(&id)
	mib.tables[&id] = &table

	return &table
}

func (mib *MIB) ResolveName(name string) (ID, error) {
	if id := mib.registry.getName(name); id == nil {
		return ID{MIB: mib, Name: name}, fmt.Errorf("%v name not found: %v", mib.Name, name)
	} else {
		return *id, nil
	}
}

func (mib *MIB) Lookup(oid snmp.OID) ID {
	if id := mib.registry.getOID(oid); id == nil {
		return ID{MIB: mib, OID: oid}
	} else {
		return *id
	}
}

func (mib *MIB) LookupObject(oid snmp.OID) *Object {
	if id := mib.registry.getOID(oid); id == nil {
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
		return fmt.Sprintf("%s%s", mib.Name, snmp.OID(index).String())
	}
}
