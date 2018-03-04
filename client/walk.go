package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

// split into multipl GetNext requests of options.MaxVars
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

// Only walks over VarBinds that are within the requested OID prefix.
// Any response VarBind outside of the requested tree is substituited with a snmp.EndOfMibView error
// Stops walking once all varbinds are done
// Splits into multiple walk requests if the number of OIDs exceeds options.MaxVars
func (client *Client) Walk(walkFunc func(...snmp.VarBind) error, startOIDs ...snmp.OID) error {
	var oids = make([]snmp.OID, len(startOIDs))

	for i, oid := range startOIDs {
		oids[i] = oid
	}

	for {
		// request splitting
		varBinds, err := client.walkNext(oids...)
		if err != nil {
			return err
		}

		// omit vars that walked out of the table
		var count = 0

		for i, varBind := range varBinds {
			oid := varBind.OID()

			if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
				// explicit SNMPv2 break
				continue
			} else if oid.Equals(oids[i]) || startOIDs[i].Index(oid) == nil {
				// not making progress, or walked out of tree
				varBinds[i] = snmp.MakeVarBind(oids[i], snmp.EndOfMibViewValue)
			} else {
				oids[i] = oid
				count++
			}
		}

		if count == 0 {
			// no valid responses
			break
		} else if err := walkFunc(varBinds...); err != nil {
			return err
		}
	}

	return nil
}
