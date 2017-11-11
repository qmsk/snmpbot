package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

func makeRegistry() registry {
	return registry{
		byOID: make(map[string]*ID),
		byName: make(map[string]*ID),
	}
}

type registry struct {
	byOID  map[string]*ID
	byName map[string]*ID
}

func (registry *registry) registerOID(id *ID) {
	registry.byOID[id.OID.String()] = id
}

func (registry *registry) registerName(id *ID) {
	registry.byName[id.Name] = id
}

func (registry *registry) register(id *ID) {
	registry.registerOID(id)
	registry.registerName(id)
}

func (registry *registry) getName(name string) *ID {
	return registry.byName[name]
}

func (registry *registry) getOID(oid snmp.OID) *ID {
	var lookup = ""

	for _, id := range oid {
		lookup += fmt.Sprintf(".%d", id)

		if id := registry.byOID[lookup]; id != nil {
			return id
		}
	}

	return nil
}
