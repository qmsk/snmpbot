package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

// Split request OIDs into multiple GetNext requests of options.MaxVars each.
//
// Override response varbinds outside of rootOIDs with snmp.EndOfMibViewValue
//
// TODO: automatically handle snmp.TooBigError?
func (client *Client) WalkNext(rootOIDs []snmp.OID, walkOIDs []snmp.OID) ([]snmp.VarBind, error) {
	var maxVars = DefaultMaxVars
	var retVars = make([]snmp.VarBind, len(walkOIDs))
	var retLen = uint(0)

	if client.options.MaxVars > 0 {
		maxVars = client.options.MaxVars
	}

	for retLen < uint(len(walkOIDs)) {
		var reqOffset = retLen
		var reqOIDs = make([]snmp.OID, maxVars)
		var reqLen = uint(0)

		for retLen+reqLen < uint(len(walkOIDs)) && reqLen < maxVars {
			reqOIDs[reqLen] = walkOIDs[reqOffset+reqLen]
			reqLen++
		}

		if varBinds, err := client.GetNext(reqOIDs[:reqLen]...); err != nil {
			return nil, err
		} else {
			for i, varBind := range varBinds {
				var oid = varBind.OID()
				var rootOID = rootOIDs[reqOffset+uint(i)]

				if oid.Equals(reqOIDs[i]) || rootOID.Index(oid) == nil {
					// not making progress, or walked out of tree
					varBinds[i] = snmp.MakeVarBind(rootOID, snmp.EndOfMibViewValue)
				}
			}

			for _, varBind := range varBinds {
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
// Returns if none of the entry varBinds are within the requested OIDs.
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

	for {
		// request splitting
		varBinds, err := client.WalkNext(rootOIDs, walkOIDs)
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

		if entryCount > 0 {
			if err := walkFunc(scalarVars, entryVars); err != nil {
				return err
			}
		}

		if entryCount == 0 {
			// not making any progress
			break
		}
	}

	return nil
}

func (client *Client) Walk(oids []snmp.OID, walkFunc func(varBinds []snmp.VarBind) error) error {
	return client.WalkWithScalars(nil, oids, func(scalars []snmp.VarBind, entries []snmp.VarBind) error {
		return walkFunc(entries)
	})
}
