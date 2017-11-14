package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
)

type hostID string

type hosts map[hostID]*Host

func (hosts hosts) makeAPIIndex() []api.HostIndex {
	var items = make([]api.HostIndex, 0, len(hosts))

	for _, host := range hosts {
		items = append(items, host.makeAPIIndex())
	}

	return items
}

func (hosts hosts) Index(name string) (web.Resource, error) {
	if host, ok := hosts[hostID(name)]; !ok {
		return nil, nil
	} else {
		return host, nil
	}
}

func (hosts hosts) GetREST() (web.Resource, error) {
	return hosts.makeAPIIndex(), nil
}
