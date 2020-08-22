package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type MIBs map[string]*mibs.MIB

func AllMIBs() MIBs {
	var mibMap = make(MIBs)

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		mibMap[mib.Name] = mib
	})

	return mibMap
}

func MakeMIBs(args ...*mibs.MIB) MIBs {
	var mibs = make(MIBs, len(args))

	for _, mib := range args {
		mibs[mib.Name] = mib
	}

	return mibs
}

func (mibs MIBs) Keys() []string {
	var keys = make([]string, 0, len(mibs))

	for key, _ := range mibs {
		keys = append(keys, key)
	}

	return keys
}

func (mibMap MIBs) ListIDs() []mibs.ID {
	var list = make([]mibs.ID, 0, len(mibMap))

	for _, mib := range mibMap {
		list = append(list, mib.ID)
	}

	return list
}

func (mibMap MIBs) Add(mib *mibs.MIB) {
	mibMap[mib.Name] = mib
}

func (mibMap MIBs) Objects() Objects {
	var objects = make(Objects)

	mibs.WalkObjects(func(object *mibs.Object) {
		if _, ok := mibMap[object.MIB.Name]; !ok {
			return
		}
		if object.NotAccessible {
			return
		}
		objects.add(object)
	})

	return objects
}

func (mibMap MIBs) Tables() Tables {
	var tables = make(Tables)

	mibs.WalkTables(func(table *mibs.Table) {
		if _, ok := mibMap[table.MIB.Name]; !ok {
			return
		}
		tables.add(table)
	})

	return tables
}

type mibsRoute struct {
	mibs MIBs
}

func (route mibsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return mibsView{route.mibs}, nil
	} else if mib, ok := route.mibs[name]; !ok {
		return nil, web.Errorf(404, "MIB not found: %v", name)
	} else {
		return mibView{mib}, nil
	}
}

type mibView struct {
	mib *mibs.MIB
}

func (view mibView) makeAPIIndex() api.MIBIndex {
	return api.MIBIndex{
		ID: view.mib.String(),
	}
}

func (view mibView) makeAPI() api.MIB {
	var mib = api.MIB{
		MIBIndex: view.makeAPIIndex(),
		Objects:  mibObjectsView{view.mib}.makeAPIIndex(),
		Tables:   mibTablesView{view.mib}.makeAPIIndex(),
	}

	return mib
}

func (view mibView) GetREST() (web.Resource, error) {
	return view.makeAPI(), nil
}

type mibsView struct {
	mibs MIBs
}

func (view mibsView) makeAPIIndex() []api.MIBIndex {
	var index []api.MIBIndex

	for _, mib := range view.mibs {
		index = append(index, mibView{mib}.makeAPIIndex())
	}

	return index
}

func (view mibsView) makeAPI() []api.MIB {
	var rets []api.MIB

	for _, mib := range view.mibs {
		rets = append(rets, mibView{mib}.makeAPI())
	}

	return rets
}

func (view mibsView) GetREST() (web.Resource, error) {
	return view.makeAPI(), nil
}
