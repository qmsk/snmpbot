package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
)

type hostID string

type hosts map[hostID]*Host

type hostsRoute struct {
	hosts hosts
}

func (route hostsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return hostsView{route.hosts}, nil
	} else if host, ok := route.hosts[hostID(name)]; !ok {
		return nil, nil
	} else {
		return hostRoute{host}, nil
	}
}

type hostsView struct {
	hosts hosts
}

func (view hostsView) makeAPIIndex() []api.HostIndex {
	var items = make([]api.HostIndex, 0, len(view.hosts))

	for _, host := range view.hosts {
		items = append(items, hostView{host}.makeAPIIndex())
	}

	return items
}

func (view hostsView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
}
