package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type EntrySyntax []*Object
type EntryValues []Value
type EntryMap map[IDKey]Value

func (entrySyntax EntrySyntax) OIDs() []snmp.OID {
	var oids = make([]snmp.OID, len(entrySyntax))

	for i, entry := range entrySyntax {
		oids[i] = entry.OID
	}

	return oids
}

func (entrySyntax EntrySyntax) Unpack(varBinds []snmp.VarBind) (EntryValues, error) {
	var entryValues = make(EntryValues, len(entrySyntax))

	for i, entryObject := range entrySyntax {
		var varBind = varBinds[i]

		if err := varBind.ErrorValue(); err != nil {
			// XXX: skip unsupported columns?
		}

		if index := entryObject.OID.Index(varBind.OID()); index == nil {
			return nil, fmt.Errorf("Invalid VarBind[%v] OID for %v: %v", varBind.OID(), entryObject, entryObject.OID)
		}

		if value, err := entryObject.Unpack(varBind); err != nil {
			return nil, fmt.Errorf("Invalid VarBind[%v] Value for %v: %v", varBind.OID(), entryObject, err)
		} else {
			entryValues[i] = value
		}
	}

	return entryValues, nil
}

func (entrySyntax EntrySyntax) Map(varBinds []snmp.VarBind) (EntryMap, error) {
	var entryMap = make(EntryMap)

	for i, entryObject := range entrySyntax {
		var varBind = varBinds[i]

		if err := varBind.ErrorValue(); err != nil {
			// XXX: skip unsupported columns?
		}

		if index := entryObject.OID.Index(varBind.OID()); index == nil {
			return nil, fmt.Errorf("Invalid VarBind[%v] OID for %v: %v", varBind.OID(), entryObject, entryObject.OID)
		}

		if value, err := entryObject.Unpack(varBind); err != nil {
			return nil, fmt.Errorf("Invalid VarBind[%v] Value for %v: %v", varBind.OID(), entryObject, err)
		} else {
			entryMap[entryObject.ID.Key()] = value
		}
	}

	return entryMap, nil
}

type Table struct {
	ID

	IndexSyntax IndexSyntax
	EntrySyntax EntrySyntax
}

func (table Table) EntryOIDs() []snmp.OID {
	return table.EntrySyntax.OIDs()
}

func (table Table) Unpack(varBinds []snmp.VarBind) (IndexValues, EntryValues, error) {
	if len(varBinds) != len(table.EntrySyntax) {
		return nil, nil, fmt.Errorf("Incorrect count of colums for Table<%v>: %d", table, len(varBinds))
	}

	// XXX: assuming all entry objects have the same index...
	var index = table.EntrySyntax[0].OID.Index(varBinds[0].OID())

	if entryValues, err := table.EntrySyntax.Unpack(varBinds); err != nil {
		return nil, entryValues, err
	} else if indexValues, err := table.IndexSyntax.UnpackIndex(index); err != nil {
		return indexValues, nil, err
	} else {
		return indexValues, entryValues, nil
	}
}

func (table Table) Map(varBinds []snmp.VarBind) (IndexMap, EntryMap, error) {
	if len(varBinds) != len(table.EntrySyntax) {
		return nil, nil, fmt.Errorf("Incorrect count of colums for Table<%v>: %d", table, len(varBinds))
	}

	// XXX: assuming all entry objects have the same index...
	var index = table.EntrySyntax[0].OID.Index(varBinds[0].OID())

	if entryMap, err := table.EntrySyntax.Map(varBinds); err != nil {
		return nil, nil, err
	} else if indexMap, err := table.IndexSyntax.MapIndex(index); err != nil {
		return nil, nil, err
	} else {
		return indexMap, entryMap, nil
	}
}
