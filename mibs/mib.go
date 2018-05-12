package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"regexp"
)

func makeMIB(name string, oid snmp.OID) MIB {
	return MIB{
		Name: name,
		OID:  oid,

		registry: makeRegistry(),
		objects:  make(map[IDKey]*Object),
		tables:   make(map[IDKey]*Table),
	}
}

type MIB struct {
	Name string
	OID  snmp.OID
	registry

	objects map[IDKey]*Object
	tables  map[IDKey]*Table
}

func (mib *MIB) String() string {
	return mib.Name
}

func (mib *MIB) registerObject(object Object) *Object {
	mibRegistry.registerOID(object.ID)
	mib.registry.register(object.ID)
	mib.objects[object.ID.Key()] = &object

	return &object
}

func (mib *MIB) registerTable(table Table) *Table {
	mibRegistry.registerOID(table.ID)
	mib.registry.register(table.ID)
	mib.tables[table.ID.Key()] = &table

	return &table
}

/* Resolve MIB-relative ID by human-readable name:
".1.0"
"sysDescr"
"sysDescr.0"
*/
var mibResolveRegexp = regexp.MustCompile("^([^.]+?)?([.][0-9.]+)?$")

func (mib *MIB) Resolve(name string) (ID, error) {
	var id = ID{MIB: mib, OID: mib.OID}
	var nameID, nameOID string

	if matches := mibResolveRegexp.FindStringSubmatch(name); matches == nil {
		return id, fmt.Errorf("Invalid syntax: %v", name)
	} else {
		nameID = matches[1]
		nameOID = matches[2]
	}

	if nameID == "" {

	} else if resolveID, err := mib.ResolveName(nameID); err != nil {
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

func (mib *MIB) Walk(f func(ID)) {
	mib.registry.walk(f)
}

func (mib *MIB) Object(id ID) *Object {
	if object, ok := mib.objects[id.Key()]; !ok {
		return nil
	} else {
		return object
	}
}

func (mib *MIB) ResolveObject(name string) *Object {
	if id, err := mib.Resolve(name); err != nil {
		return nil
	} else {
		return mib.Object(id)
	}
}

func (mib *MIB) Table(id ID) *Table {
	if table, ok := mib.tables[id.Key()]; !ok {
		return nil
	} else {
		return table
	}
}

func (mib *MIB) ResolveTable(name string) *Table {
	if id, err := mib.Resolve(name); err != nil {
		return nil
	} else {
		return mib.Table(id)
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
