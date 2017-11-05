package client

import (
	"github.com/qmsk/snmpbot/snmp"
)

func (client *Client) Walk(walkFunc func(...snmp.VarBind) error, startOIDs ...snmp.OID) error {
	var oids = make([]snmp.OID, len(startOIDs))

	for i, oid := range startOIDs {
		oids[i] = oid
	}

	for {
		if varBinds, err := client.GetNext(oids...); err != nil {
			return err
		} else {
			for i, varBind := range varBinds {
				oid := varBind.OID()

				if errorValue := varBind.ErrorValue(); errorValue == snmp.EndOfMibViewValue {
					// explicit SNMPv2 break
					return nil
				} else if oid.Equals(oids[i]) || startOIDs[i].Index(oid) == nil {
					// not making progress, or walked out of tree
					return nil
				} else {
					oids[i] = oid
				}
			}

			if err := walkFunc(varBinds...); err != nil {
				return err
			}
		}
	}
}
