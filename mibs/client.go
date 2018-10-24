package mibs

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/snmp"
)

func MakeClient(c *client.Client) Client {
	return Client{Client: c}
}

type Client struct {
	*client.Client
}

// Probe the MIB at id
func (client Client) Probe(ids []ID) ([]bool, error) {
	var oids = make([]snmp.OID, len(ids))
	var probed = make([]bool, len(ids))

	for i, id := range ids {
		oids[i] = id.OID
	}

	if varBinds, err := client.GetScalars(oids); err != nil {
		return probed, err
	} else {
		for i, varBind := range varBinds {
			if err := varBind.ErrorValue(); err != nil {
				// not supported
			} else {
				probed[i] = true
			}
		}
	}

	return probed, nil
}

func (client Client) WalkObjects(objects []*Object, f func(*Object, IndexValues, Value, error) error) error {
	var oids = make([]snmp.OID, len(objects))

	for i, object := range objects {
		oids[i] = object.OID
	}

	return client.Client.WalkObjects(oids, func(varBinds []snmp.VarBind) error {
		for i, varBind := range varBinds {
			var object = objects[i]
			var walkErr error

			if err := varBind.ErrorValue(); err != nil {
				// just skip unsupported objects...
			} else if value, err := object.Unpack(varBind); err != nil {
				walkErr = f(object, nil, value, err)
			} else if indexValues, err := object.UnpackIndex(varBind.OID()); err != nil {
				walkErr = f(object, indexValues, value, err)
			} else {
				walkErr = f(object, indexValues, value, nil)
			}

			if walkErr != nil {
				return walkErr
			}
		}

		return nil
	})
}

func (client Client) WalkTable(table *Table, f func(IndexValues, EntryValues, error) error) error {
	return client.Client.WalkTable(table.EntryOIDs(), func(varBinds []snmp.VarBind) error {
		indexValues, entryValues, err := table.Unpack(varBinds)

		return f(indexValues, entryValues, err)
	})
}
