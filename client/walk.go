package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

func walkScalarVars(oids []snmp.OID, varBinds []snmp.VarBind) bool {
	var ok = false

	for i, varBind := range varBinds {
		var oid = varBind.OID()

		if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
			// explicit SNMPv2 break
		} else if oid.Equals(oids[i]) || oids[i].Index(oid) == nil {
			// not making progress, or walked out of tree
			varBinds[i] = snmp.MakeVarBind(oids[i], snmp.EndOfMibViewValue)
		} else {
			ok = true
		}
	}

	return ok
}

func walkObjectVars(rootOIDs []snmp.OID, walkOIDs []snmp.OID, varBinds []snmp.VarBind) bool {
	var ok = false

	for i, varBind := range varBinds {
		var rootOID = rootOIDs[i]
		var oid = varBind.OID()

		if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
			// explicit SNMPv2 break
		} else if oid.Equals(walkOIDs[i]) || rootOID.Index(oid) == nil {
			// not making progress, or walked out of tree
			varBinds[i] = snmp.MakeVarBind(rootOID, snmp.EndOfMibViewValue)
		} else {
			walkOIDs[i] = oid
			ok = true
		}
	}

	return ok
}

func indexCmp(a []int, b []int) int {
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return +1
	}

	for i, _ := range a {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return +1
		}
	}

	return 0 // equal
}

func walkEntryVars(rootOIDs []snmp.OID, walkOIDs []snmp.OID, varBinds []snmp.VarBind) bool {
	var ok = false
	var entryIndex []int

	// select minimum index
	for i, varBind := range varBinds {
		var rootOID = rootOIDs[i]
		var oid = varBind.OID()

		if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
			// explicit SNMPv2 break
			continue
		} else if oid.Equals(walkOIDs[i]) {
			// not making progress
			continue

		} else if index := rootOID.Index(oid); index == nil {
			// walked out of tree
			varBinds[i] = snmp.MakeVarBind(rootOID, snmp.EndOfMibViewValue)
			continue

		} else if entryIndex == nil || indexCmp(entryIndex, index) > 0 {
			entryIndex = index
		}

		// at least one object is making progress
		ok = true
	}

	// select objects with matching index
	for i, varBind := range varBinds {
		var rootOID = rootOIDs[i]
		var oid = varBind.OID()
		var entryOID = rootOID.Extend(entryIndex...)

		if index := rootOID.Index(oid); index == nil {
			// error, leave as-is
		} else if indexCmp(entryIndex, index) != 0 {
			// hole, replace with snmp.NoSuchInstanceValue
			varBinds[i] = snmp.MakeVarBind(entryOID, snmp.NoSuchInstanceValue)
		} else {
			// valid, leave as-is
		}

		walkOIDs[i] = entryOID
	}

	return ok
}

type WalkOptions struct {
	// scalar objects with only one instance (.0), each walk step returns the same object instance
	Scalars []snmp.OID

	// mixed objects, each walk step may return objects with different indexes
	Objects []snmp.OID

	// table entry objects, each walk step returns objects with the same index
	TableEntries []snmp.OID
}

type WalkFunc func(vars []snmp.VarBind) error

// Object/Table traversal using GetNext.
//
// Scalar OIDs are objects that are always fetched for every traversed row.
// Entry OIDs are objects that are fetched based on the previous row.
//
// Only returns VarBinds that are within the requested OID prefixes.
// Any response VarBind outside of the requested tree is substituited with a snmp.EndOfMibView error value.
//
// Returns if none of the entry varBinds are within the requested OIDs.
//
// Splits into multiple requests if the number of OIDs exceeds options.MaxVars.
func (client *Client) WalkWithOptions(options WalkOptions, walkFunc WalkFunc) error {
	if client.options.NoBulk {
		return client.walkGetNext(options, walkFunc)
	} else {
		return client.walkGetBulk(options, walkFunc)
	}
}

func (client *Client) walkGetNext(options WalkOptions, walkFunc WalkFunc) error {
	var walkOIDs = make([]snmp.OID, len(options.Scalars)+len(options.Objects)+len(options.TableEntries))
	var objectsOffset = len(options.Scalars)
	var entriesOffset = len(options.Scalars) + len(options.Objects)

	for i, oid := range options.Scalars {
		walkOIDs[i] = oid
	}
	for i, oid := range options.Objects {
		walkOIDs[objectsOffset+i] = oid
	}
	for i, oid := range options.TableEntries {
		walkOIDs[entriesOffset+i] = oid
	}

	if len(walkOIDs) == 0 {
		return nil
	}

	for {
		// request splitting
		varBinds, err := client.GetNextSplit(walkOIDs)
		if err != nil {
			return err
		}

		if !walkScalarVars(options.Scalars, varBinds[0:len(options.Scalars)]) {
			// no scalar vars matched !?
		}

		if len(options.Objects) > 0 && !walkObjectVars(options.Objects, walkOIDs[objectsOffset:objectsOffset+len(options.Objects)], varBinds[objectsOffset:objectsOffset+len(options.Objects)]) {
			// did not make progress
			return nil
		}

		if len(options.TableEntries) > 0 && !walkEntryVars(options.TableEntries, walkOIDs[entriesOffset:entriesOffset+len(options.TableEntries)], varBinds[entriesOffset:entriesOffset+len(options.TableEntries)]) {
			// did not make progress
			return nil
		}

		if err := walkFunc(varBinds); err != nil {
			return err
		}

		if len(options.Objects) == 0 && len(options.TableEntries) == 0 {
			return nil
		}
	}
}

func (client *Client) walkGetBulk(options WalkOptions, walkFunc WalkFunc) error {
	var walkOIDs = make([]snmp.OID, len(options.Objects)+len(options.TableEntries))
	var entriesOffset = len(options.Objects)

	for i, oid := range options.Objects {
		walkOIDs[i] = oid
	}
	for i, oid := range options.TableEntries {
		walkOIDs[entriesOffset+i] = oid
	}

	for {
		// TODO: request splitting
		scalarVars, entryList, err := client.GetBulk(options.Scalars, walkOIDs)
		if err != nil {
			return err
		}

		if !walkScalarVars(options.Scalars, scalarVars) {
			// no scalar vars matched !?
		}

		var ok = false

		for _, entryVars := range entryList {
			if len(options.Objects) > 0 && !walkObjectVars(options.Objects, walkOIDs[0:len(options.Objects)], entryVars[0:len(options.Objects)]) {
				// no vars made progress, ignore the remainder
				ok = false
				break
			}

			if len(options.TableEntries) > 0 && !walkEntryVars(options.TableEntries, walkOIDs[entriesOffset:entriesOffset+len(options.TableEntries)], entryVars[entriesOffset:entriesOffset+len(options.TableEntries)]) {
				// no vars made progress, ignore the remainder
				ok = false
				break
			}

			var vars = make([]snmp.VarBind, len(options.Scalars)+len(options.Objects)+len(options.TableEntries))

			for i, v := range scalarVars {
				vars[i] = v
			}
			for i, v := range entryVars {
				vars[len(scalarVars)+i] = v
			}

			if err := walkFunc(vars); err != nil {
				return err
			} else {
				ok = true
			}
		}

		if !ok {
			return nil
		}
	}
}

// Perform a single GetNext walk step, returning either objects underneath given oid, or EndOfMibViewValue
func (client *Client) GetScalars(oids []snmp.OID) ([]snmp.VarBind, error) {
	var retVars []snmp.VarBind

	// walkGetBulk is useless, and doesn't support scalars-only
	return retVars, client.walkGetNext(WalkOptions{Scalars: oids}, func(vars []snmp.VarBind) error {
		retVars = vars

		return nil
	})
}

func (client *Client) WalkObjects(oids []snmp.OID, walkFunc WalkFunc) error {
	return client.WalkWithOptions(WalkOptions{Objects: oids}, walkFunc)
}

func (client *Client) WalkTable(entryOids []snmp.OID, walkFunc WalkFunc) error {
	return client.WalkWithOptions(WalkOptions{TableEntries: entryOids}, walkFunc)
}
