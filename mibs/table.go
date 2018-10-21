package mibs

import (
	"fmt"
	"strings"

	"github.com/qmsk/snmpbot/snmp"
)

type EntrySyntax []*Object
type EntryValues []Value
type EntryMap map[IDKey]Value

type EntryErrors []error

func (errs *EntryErrors) add(err error) {
	*errs = append(*errs, err)
}

func (errs EntryErrors) Error() string {
	var strs = make([]string, len(errs))

	for i, err := range errs {
		strs[i] = err.Error()
	}

	return strings.Join(strs, "; ")
}

func (entrySyntax EntrySyntax) OIDs() []snmp.OID {
	var oids = make([]snmp.OID, len(entrySyntax))

	for i, entry := range entrySyntax {
		oids[i] = entry.OID
	}

	return oids
}

func indexEquals(expected []int, index []int) bool {
	if len(expected) != len(index) {
		return false
	}

	for i, x := range expected {
		if index[i] != x {
			return false
		}
	}

	return true
}

// Returns nil entries for objects with error values
func (entrySyntax EntrySyntax) Unpack(varBinds []snmp.VarBind) ([]int, EntryValues, error) {
	var entryValues = make(EntryValues, len(entrySyntax))
	var entryIndex []int
	var entryErrors EntryErrors

	if len(varBinds) != len(entrySyntax) {
		return nil, nil, fmt.Errorf("Invalid VarBinds[%v] for entry syntax: %v", varBinds, entrySyntax)
	}

	for i, entryObject := range entrySyntax {
		var varBind = varBinds[i]

		if err := varBind.ErrorValue(); err != nil {
			// skip unsupported columns
		} else if index := entryObject.OID.Index(varBind.OID()); index == nil {
			entryErrors.add(fmt.Errorf("Invalid VarBind[%v] OID for %v: %v", varBind.OID(), entryObject, entryObject.OID))
		} else if entryIndex != nil && !indexEquals(entryIndex, index) {
			entryErrors.add(fmt.Errorf("Mismatching VarBind[%v] OID for %v: index %v != expected %v", varBind.OID(), entryObject, index, entryIndex))
		} else if value, err := entryObject.Unpack(varBind); err != nil {
			entryErrors.add(fmt.Errorf("Invalid VarBind[%v] Value for %v: %v", varBind.OID(), entryObject, err))
		} else {
			entryIndex = index
			entryValues[i] = value
		}
	}

	if entryErrors == nil {
		// interface with type but nil value does not compare equal to nil
		return entryIndex, entryValues, nil
	} else {
		return entryIndex, entryValues, entryErrors
	}
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
	index, entryValues, entryErr := table.EntrySyntax.Unpack(varBinds)

	if index == nil {
		return nil, entryValues, entryErr
	}

	indexValues, indexErr := table.IndexSyntax.UnpackIndex(index)

	if indexErr != nil {
		return indexValues, entryValues, indexErr
	} else if entryErr != nil {
		return indexValues, entryValues, entryErr
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
