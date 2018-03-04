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

func walkEntryVars(rootOIDs []snmp.OID, walkOIDs []snmp.OID, varBinds []snmp.VarBind) bool {
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
func (client *Client) WalkWithScalars(scalars []snmp.OID, entries []snmp.OID, walkFunc func(scalars []snmp.VarBind, entries []snmp.VarBind) error) error {
	if client.options.NoBulk {
		return client.walkGetNext(scalars, entries, walkFunc)
	} else {
		return client.walkGetBulk(scalars, entries, walkFunc)
	}
}

func (client *Client) walkGetNext(scalars []snmp.OID, entries []snmp.OID, walkFunc func(scalars []snmp.VarBind, entries []snmp.VarBind) error) error {
	var walkOIDs = make([]snmp.OID, len(scalars)+len(entries))

	for i, oid := range scalars {
		walkOIDs[i] = oid
	}
	for i, oid := range entries {
		walkOIDs[len(scalars)+i] = oid
	}

	for {
		// request splitting
		varBinds, err := client.GetNextSplit(walkOIDs)
		if err != nil {
			return err
		}

		var scalarVars = varBinds[:len(scalars)]
		var entryVars = varBinds[len(scalars):]

		if !walkScalarVars(scalars, scalarVars) {
			// no scalar vars matched !?
		}

		if !walkEntryVars(entries, walkOIDs[len(scalars):], entryVars) {
			// did not make progress
			return nil
		}

		if err := walkFunc(scalarVars, entryVars); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) walkGetBulk(scalars []snmp.OID, entries []snmp.OID, walkFunc func(scalars []snmp.VarBind, entries []snmp.VarBind) error) error {
	var walkOIDs = make([]snmp.OID, len(entries))

	for i, oid := range entries {
		walkOIDs[i] = oid
	}

	for {
		// TODO: request splitting
		scalarVars, entryList, err := client.GetBulk(scalars, walkOIDs)
		if err != nil {
			return err
		}

		if !walkScalarVars(scalars, scalarVars) {
			// no scalar vars matched !?
		}

		for _, entryVars := range entryList {
			if !walkEntryVars(entries, walkOIDs, entryVars) {
				// did not make progress
				return nil
			}

			if err := walkFunc(scalarVars, entryVars); err != nil {
				return err
			}
		}
	}
}

func (client *Client) Walk(oids []snmp.OID, walkFunc func(varBinds []snmp.VarBind) error) error {
	return client.WalkWithScalars(nil, oids, func(scalars []snmp.VarBind, entries []snmp.VarBind) error {
		return walkFunc(entries)
	})
}
