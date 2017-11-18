package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type mibsRoute struct {
}

func (route mibsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return mibsView{}, nil
	} else if mib, err := mibs.ResolveMIB(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
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
}

func (_ mibsView) makeAPIIndex() []api.MIBIndex {
	var index []api.MIBIndex

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		index = append(index, mibView{mib}.makeAPIIndex())
	})

	return index
}

func (_ mibsView) makeAPI() []api.MIB {
	var rets []api.MIB

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		rets = append(rets, mibView{mib}.makeAPI())
	})

	return rets
}

func (view mibsView) GetREST() (web.Resource, error) {
	return view.makeAPI(), nil
}
