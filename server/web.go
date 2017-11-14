package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
)

func (engine *Engine) WebAPI() web.API {
	return web.MakeAPI(engine)
}

func (engine *Engine) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return engine, nil
	case "hosts":
		return engine.hosts, nil
	default:
		return nil, nil
	}
}

func (engine *Engine) GetREST() (web.Resource, error) {
	var index = api.Index{
		MIBs:  engine.mibRegistry.makeAPIIndex(),
		Hosts: engine.hosts.makeAPIIndex(),
	}

	return index, nil
}
