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
	var index = api.MIBIndex{
		ID:      view.mib.String(),
		Objects: []api.ObjectIndex{},
		Tables:  []api.TableIndex{},
	}

	view.mib.Walk(func(id mibs.ID) {
		if object := view.mib.Object(id); object != nil {
			index.Objects = append(index.Objects, objectView{object}.makeAPIIndex())
		}

		if table := view.mib.Table(id); table != nil {
			index.Tables = append(index.Tables, tableView{table}.makeAPIIndex())
		}
	})

	return index
}

func (view mibView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
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

func (view mibsView) GetREST() (web.Resource, error) {
	return view.makeAPIIndex(), nil
}
