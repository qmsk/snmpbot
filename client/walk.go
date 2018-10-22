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

type WalkOptions struct {
	// scalar objects with only one instance (.0), each walk step returns the same object instance
	Scalars []snmp.OID

	// mixed objects, each walk step may return objects with different indexes
	Objects []snmp.OID
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
	var walkOIDs = make([]snmp.OID, len(options.Scalars)+len(options.Objects))

	for i, oid := range options.Scalars {
		walkOIDs[i] = oid
	}
	for i, oid := range options.Objects {
		walkOIDs[len(options.Scalars)+i] = oid
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

		if !walkObjectVars(options.Objects, walkOIDs[len(options.Scalars):len(options.Scalars)+len(options.Objects)], varBinds[len(options.Scalars):len(options.Scalars)+len(options.Objects)]) {
			// did not make progress
			return nil
		}

		if err := walkFunc(varBinds); err != nil {
			return err
		}
	}
}

func (client *Client) walkGetBulk(options WalkOptions, walkFunc WalkFunc) error {
	var walkOIDs = make([]snmp.OID, len(options.Objects))

	for i, oid := range options.Objects {
		walkOIDs[i] = oid
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
			if !walkObjectVars(options.Objects, walkOIDs[0:len(options.Objects)], entryVars[0:len(options.Objects)]) {
				// no vars made progress, ignore the remainder
				ok = false
				break
			}

			var vars = make([]snmp.VarBind, len(options.Scalars)+len(options.Objects))

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

	return retVars, client.WalkWithOptions(WalkOptions{Objects: oids}, func(vars []snmp.VarBind) error {
		retVars = vars

		return nil
	})
}

func (client *Client) WalkObjects(oids []snmp.OID, walkFunc WalkFunc) error {
	return client.WalkWithOptions(WalkOptions{Objects: oids}, walkFunc)
}
