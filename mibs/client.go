package mibs

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/snmp"
)

type Client struct {
	*client.Client
}

// Read the value at object index .0
func (client *Client) GetObject(object *Object) (Value, error) {
	if varBinds, err := client.Get(object.OID.Extend(0)); err != nil {
		return nil, err
	} else if value, err := object.Unpack(varBinds[0]); err != nil {
		return nil, err
	} else {
		return value, nil
	}
}

func (client *Client) WalkTable(table *Table, f func(IndexMap, EntryMap) error) error {
	return client.Walk(func(varBinds ...snmp.VarBind) error {
		if indexMap, entryMap, err := table.Map(varBinds); err != nil {
			return err
		} else if err := f(indexMap, entryMap); err != nil {
			return err
		} else {
			return nil
		}
	}, table.EntrySyntax.OIDs()...)
}
