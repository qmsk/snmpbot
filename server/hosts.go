package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"path"
	"strings"
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

func (hosts Hosts) String() string {
	var ss = make([]string, 0, len(hosts))

	for _, host := range hosts {
		ss = append(ss, host.String())
	}

	return "{" + strings.Join(ss, ", ") + "}"
}

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
	engine    Engine
	hosts     Hosts
	hostQuery api.HostQuery
}

func (route *hostsRoute) QueryREST() interface{} {
	return &route.hostQuery
}

func (route *hostsRoute) makeHostConfig() HostConfig {
	var options = route.engine.ClientOptions()

	if route.hostQuery.Community != "" {
		options.Community = route.hostQuery.Community
	}

	return HostConfig{
		SNMP:          route.hostQuery.SNMP,
		ClientOptions: &options,
	}
}

func (route *hostsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return &hostsView{engine: route.engine, hosts: route.hosts}, nil
	} else if host, ok := route.hosts[HostID(name)]; ok {
		return &hostRoute{engine: route.engine, host: host}, nil
	} else {
		var host = newHost(HostID(name))
		var hostConfig = route.makeHostConfig()

		return &hostRoute{
			engine:     route.engine,
			host:       host,
			loadConfig: &hostConfig, // apply at route lookup
		}, nil
	}
}

type hostsView struct {
	engine Engine
	hosts  Hosts
	post   api.HostPOST
}

func (view hostsView) makeAPIIndex() []api.HostIndex {
	var items = make([]api.HostIndex, 0, len(view.hosts))

	for _, host := range view.hosts {
		items = append(items, hostView{host: host}.makeAPIIndex())
	}

	return items
}

func (view *hostsView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
}

func (view *hostsView) IntoREST() interface{} {
	return &view.post
}

func (view *hostsView) makeHostConfig() HostConfig {
	var options = view.engine.ClientOptions()

	if view.post.Community != "" {
		options.Community = view.post.Community
	}

	return HostConfig{
		SNMP:          view.post.SNMP,
		Location:      view.post.Location,
		ClientOptions: &options,
	}
}

func (view *hostsView) PostREST() (web.Resource, error) {
	if host, err := loadHost(view.engine, HostID(view.post.ID), view.makeHostConfig()); err != nil {
		return nil, err
	} else if ok := view.engine.AddHost(host); !ok {
		return nil, web.Errorf(409, "Host already configured: %v", host.id)
	} else {
		return hostView{host: host}.makeAPIIndex(), nil
	}
}
