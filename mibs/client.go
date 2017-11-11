package mibs

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/snmp"
)

type Client struct {
  *client.Client
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
