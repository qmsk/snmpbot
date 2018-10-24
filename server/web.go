package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
)

func WebAPI(engine Engine) web.API {
	return web.MakeAPI(indexRoute{engine})
}

type indexRoute struct {
	engine Engine
}

func (route indexRoute) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return indexView{route.engine}, nil
	case "mibs":
		return mibsRoute{route.engine.MIBs()}, nil
	case "objects":
		return objectsRoute{route.engine}, nil
	case "tables":
		return tablesRoute{route.engine}, nil
	case "hosts":
		return &hostsRoute{engine: route.engine, hosts: route.engine.Hosts()}, nil
	default:
		return nil, nil
	}
}

type indexView struct {
	engine Engine
}

func (view indexView) makeAPIIndex() api.Index {
	return api.Index{
		MIBs:         mibsView{view.engine.MIBs()}.makeAPIIndex(),
		IndexObjects: objectsRoute{}.makeIndex(),
		IndexTables:  tablesRoute{}.makeIndex(),
		Hosts:        hostsView{hosts: view.engine.Hosts()}.makeAPIIndex(),
	}
}

func (view indexView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
}
