package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

func makeMIB(id ID) MIB {
	return MIB{
		ID:       id,
		registry: makeRegistry(),
		objects:  make(map[IDKey]*Object),
		tables:   make(map[IDKey]*Table),
	}
}

type MIB struct {
	ID
	registry

	objects map[IDKey]*Object
	tables  map[IDKey]*Table
}

func (mib *MIB) String() string {
	return mib.OID.String()
}

func (mib *MIB) MakeID(name string, ids ...int) ID {
	return ID{mib, name, mib.OID.Extend(ids...)}
}

func (mib *MIB) RegisterObject(id ID, object Object) *Object {
	object.ID = id

	mib.registry.registerOID(id)
	mib.registry.registerName(id)

	mib.objects[id.Key()] = &object

	return &object
}

func (mib *MIB) RegisterTable(id ID, table Table) *Table {
	table.ID = id

	mib.registry.registerName(id)
	mib.tables[id.Key()] = &table

	return &table
}

func (mib *MIB) ResolveName(name string) (ID, error) {
	if id, ok := mib.registry.getName(name); !ok {
		return ID{MIB: mib, Name: name}, fmt.Errorf("%v name not found: %v", mib.Name, name)
	} else {
		return id, nil
	}
}

func (mib *MIB) Lookup(oid snmp.OID) ID {
	if id, ok := mib.registry.getOID(oid); !ok {
		return ID{MIB: mib, OID: oid}
	} else {
		return id
	}
}

func (mib *MIB) Object(id ID) *Object {
	if object, ok := mib.objects[id.Key()]; !ok {
		return nil
	} else {
		return object
	}
}

func (mib *MIB) Table(id ID) *Table {
	if table, ok := mib.tables[id.Key()]; !ok {
		return nil
	} else {
		return table
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
