package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

// split into multiple GetNext requests of options.MaxVars
//
// TODO: automatically handle snmp.TooBigError?
func (client *Client) walkNext(oids ...snmp.OID) ([]snmp.VarBind, error) {
	var maxVars = DefaultMaxVars
	var retVars = make([]snmp.VarBind, len(oids))
	var retLen = uint(0)

	if client.options.MaxVars > 0 {
		maxVars = client.options.MaxVars
	}

	for retLen < uint(len(oids)) {
		var reqOIDs = make([]snmp.OID, maxVars)
		var reqLen = uint(0)

		for retLen+reqLen < uint(len(oids)) && reqLen < maxVars {
			reqOIDs[reqLen] = oids[retLen+reqLen]
			reqLen++
		}

		if getVars, err := client.GetNext(reqOIDs[:reqLen]...); err != nil {
			return nil, err
		} else {
			for _, varBind := range getVars {
				retVars[retLen] = varBind
				retLen++
			}
		}
	}

	return retVars, nil
}

// Object/Table traversal using GetNext.
//
// Scalar OIDs are objects that are always fetched for every traversed row.
// Entry OIDs are objects that are fetched based on the previous row.
//
// Only returns VarBinds that are within the requested OID prefixes.
// Any response VarBind outside of the requested tree is substituited with a snmp.EndOfMibView error value.
//
// Returns once all entry varBinds are outside of the requested OIDs.
// Exception: walking without any entry OIDs will walk the scalar OIDs exactly once.
//
// Splits into multiple requests if the number of OIDs exceeds options.MaxVars.
func (client *Client) WalkWithScalars(scalars []snmp.OID, entries []snmp.OID, walkFunc func(scalars []snmp.VarBind, entries []snmp.VarBind) error) error {
	var rootOIDs = make([]snmp.OID, len(scalars)+len(entries))
	var walkOIDs = make([]snmp.OID, len(scalars)+len(entries))
	var entryOffset = len(scalars)

	for i, oid := range scalars {
		rootOIDs[i] = oid
		walkOIDs[i] = oid
	}
	for i, oid := range entries {
		rootOIDs[entryOffset+i] = oid
		walkOIDs[entryOffset+i] = oid
	}

	// count scalar vars that we have walked
	var scalarCount = 0

	for {
		// request splitting
		varBinds, err := client.walkNext(walkOIDs...)
		if err != nil {
			return err
		}

		var scalarVars = make([]snmp.VarBind, len(scalars))
		var entryVars = make([]snmp.VarBind, len(entries))

		// count entry vars that made progress
		var entryCount = 0

		for i, varBind := range varBinds {
			oid := varBind.OID()

			if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
				// explicit SNMPv2 break
			} else if oid.Equals(walkOIDs[i]) || rootOIDs[i].Index(oid) == nil {
				// not making progress, or walked out of tree
				varBind = snmp.MakeVarBind(rootOIDs[i], snmp.EndOfMibViewValue)
			} else if i >= len(scalars) {
				// making progress on entry objects
				entryCount++
				walkOIDs[i] = oid
			}

			if i >= entryOffset {
				entryVars[i-entryOffset] = varBind
			} else {
				scalarVars[i] = varBind
			}
		}

		if entryCount > 0 || (scalarCount == 0 && len(entries) == 0) {
			if err := walkFunc(scalarVars, entryVars); err != nil {
				return err
			}
		}

		if entryCount == 0 {
			// not making any progress
			break
		}

		scalarCount++
	}

	return nil
}

func (client *Client) Walk(oids []snmp.OID, walkFunc func(varBinds []snmp.VarBind) error) error {
	return client.WalkWithScalars(nil, oids, func(scalars []snmp.VarBind, entries []snmp.VarBind) error {
		return walkFunc(entries)
	})
}
