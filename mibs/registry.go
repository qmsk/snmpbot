package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

func makeRegistry() registry {
	return registry{
		byOID:  make(map[string]ID),
		byName: make(map[string]ID),
	}
}

type registry struct {
	byOID  map[string]ID
	byName map[string]ID
}

func (registry *registry) registerOID(id ID) {
	registry.byOID[id.OID.String()] = id
}
func (registry *registry) registerName(id ID, name string) {
	registry.byName[name] = id
}

func (registry *registry) register(id ID) {
	registry.registerOID(id)
	registry.registerName(id, id.Name)
}

func (registry *registry) getName(name string) (ID, bool) {
	if id, ok := registry.byName[name]; !ok {
		return ID{Name: name}, false
	} else {
		return id, true
	}
}

func (registry *registry) getOID(oid snmp.OID) (ID, bool) {
	var key = ""
	var id = ID{OID: oid}
	var ok = false

	for _, x := range oid {
		key += fmt.Sprintf(".%d", x)

		if getID, getOK := registry.byOID[key]; getOK {
			id = getID
			ok = true
		}
	}

	return id, ok
}

func (registry *registry) walk(f func(ID)) {
	for _, id := range registry.byOID {
		f(id)
	}
}
