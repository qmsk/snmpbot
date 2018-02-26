package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"path"
)

type HostID string

func MakeHosts(args ...*Host) Hosts {
	var hosts = make(Hosts, len(args))

	for _, host := range args {
		hosts[host.id] = host
	}

	return hosts
}

type Hosts map[HostID]*Host

func (hosts Hosts) Filter(filters ...string) Hosts {
	var filtered = make(Hosts)

	for hostID, host := range hosts {
		var match = false
		var name = host.String()

		for _, filter := range filters {
			if matched, _ := path.Match(filter, name); matched {
				match = true
			}
		}

		if match {
			filtered[hostID] = host
		}
	}

	return filtered
}

type hostsRoute struct {
	engine    *Engine
	hosts     Hosts
	hostQuery api.HostQuery
}

func (route *hostsRoute) QueryREST() interface{} {
	return &route.hostQuery
}

func (route *hostsRoute) hostConfig() HostConfig {
	var options = route.engine.clientDefaults

	if route.hostQuery.Community != "" {
		options.Community = route.hostQuery.Community
	}

	return HostConfig{
		Host: route.hostQuery.Host,
		SNMP: &options,
	}
}

func (route *hostsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return hostsView{route.hosts}, nil
	} else if host, ok := route.hosts[HostID(name)]; ok {
		return hostRoute{route.engine, host}, nil
	} else {
		if host, err := newHost(route.engine, HostID(name), route.hostConfig()); err != nil {
			return nil, err
		} else if err := host.probe(); err != nil {
			return nil, err
		} else {
			return hostRoute{route.engine, host}, nil
		}
	}
}

type hostsView struct {
	hosts Hosts
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
