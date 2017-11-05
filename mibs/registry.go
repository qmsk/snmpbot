package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

type Registry struct {
	lookup  map[string]*MIB
	resolve map[string]*MIB
}

func (registry *Registry) Register(mib *MIB) {
	registry.lookup[mib.OID.String()] = mib
	registry.resolve[mib.Name] = mib
}

func (registry *Registry) Resolve(name string) *MIB {
	return registry.resolve[name]
}

func (registry *Registry) Lookup(oid snmp.OID) *MIB {
	var lookup = ""

	for _, id := range oid {
		lookup += fmt.Sprintf(".%d", id)

		if mib := registry.lookup[lookup]; mib != nil {
			return mib
		}
	}

	return nil
}
