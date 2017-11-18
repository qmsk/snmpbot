package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

// Only walks over VarBinds that are within the requested OID prefix.
// Any response VarBind outside of the requested tree is substituited with a snmp.EndOfMibView error
// Stops walking once all varbinds are done
func (client *Client) Walk(walkFunc func(...snmp.VarBind) error, startOIDs ...snmp.OID) error {
	var oids = make([]snmp.OID, len(startOIDs))

	for i, oid := range startOIDs {
		oids[i] = oid
	}

	for {
		if varBinds, err := client.GetNext(oids...); err != nil {
			return err
		} else {
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
	}

	return nil
}
