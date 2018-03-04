package mibs

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/snmp"
)

type Client struct {
	*client.Client
}

// Probe the MIB at id
func (client Client) Probe(id ID) (bool, error) {
	if varBinds, err := client.GetNext(id.OID); err != nil {
		return false, err
	} else if index := id.OID.Index(varBinds[0].OID()); index == nil {
		return false, err
	} else {
		return true, nil
	}
}

func (client Client) ProbeMany(ids []ID) (map[IDKey]bool, error) {
	var oids = make([]snmp.OID, len(ids))
	var probed = make(map[IDKey]bool)

	for i, id := range ids {
		oids[i] = id.OID
	}

	// XXX: limit on the number of var-binds per request?
	if varBinds, err := client.GetNext(oids...); err != nil {
		return probed, err
	} else {
		for i, varBind := range varBinds {
			id := ids[i]
			index := id.OID.Index(varBind.OID())

			probed[id.Key()] = (index != nil)
		}
	}

	return probed, nil
}

func (client Client) WalkObjects(objects []*Object, f func(*Object, IndexValues, Value, error) error) error {
	var oids = make([]snmp.OID, len(objects))

	for i, object := range objects {
		oids[i] = object.OID
	}

	return client.Walk(oids, func(varBinds []snmp.VarBind) error {
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

func (client Client) WalkTable(table *Table, f func(IndexValues, EntryValues) error) error {
	return client.Walk(table.EntryOIDs(), func(varBinds []snmp.VarBind) error {
		if indexValues, entryValues, err := table.Unpack(varBinds); err != nil {
			return err
		} else if err := f(indexValues, entryValues); err != nil {
			return err
		} else {
			return nil
		}
	})
}
