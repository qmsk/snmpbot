package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

// Only walks over VarBinds that are within the requested OID prefix.
// May walk with fewer VarBinds if some of them do not have any sub-objects.
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
			var valid = make([]snmp.VarBind, 0, len(varBinds))

			for i, varBind := range varBinds {
				oid := varBind.OID()

				if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
					// explicit SNMPv2 break
					continue
				} else if oid.Equals(oids[i]) || startOIDs[i].Index(oid) == nil {
					// not making progress, or walked out of tree
					continue
				} else {
					oids[i] = oid
					valid = append(valid, varBind)
				}
			}

			if len(valid) == 0 {
				break
			} else if err := walkFunc(valid...); err != nil {
				return err
			}
		}
	}

	return nil
}
