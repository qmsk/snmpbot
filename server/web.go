package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
)

func (engine *Engine) WebAPI() web.API {
	return web.MakeAPI(indexRoute{engine})
}

type indexRoute struct {
	engine *Engine
}

func (route indexRoute) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return indexView{route.engine}, nil
	case "mibs":
		return mibsRoute{}, nil
	case "objects":
		return objectsRoute{route.engine}, nil
	case "tables":
		return tablesRoute{}, nil
	case "hosts":
		return hostsRoute{route.engine, route.engine.hosts}, nil
	default:
		return nil, nil
	}
}

type indexView struct {
	engine *Engine
}

func (view indexView) makeAPIIndex() api.Index {
	return api.Index{
		MIBs:         mibsView{}.makeAPIIndex(),
		IndexObjects: objectsRoute{}.makeIndex(),
		IndexTables:  tablesRoute{}.makeIndex(),
		Hosts:        hostsView{view.engine.hosts}.makeAPIIndex(),
	}
}

func (view indexView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
}
