package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type objectView struct {
	*mibs.Object
}

func (view objectView) makeAPIIndex() api.ObjectIndex {
	var index = api.ObjectIndex{
		ID: view.Object.String(),
	}

	return index
}

type objectsView struct{}

func (view objectsView) makeAPIIndex() []api.ObjectIndex {
	var objects []api.ObjectIndex

	mibs.Walk(func(id mibs.ID) {
		if object := id.MIB.Object(id); object != nil {
			objects = append(objects, objectView{object}.makeAPIIndex())
		}
	})

	return objects
}

type mibObjectsView struct {
	mib *mibs.MIB
}

func (view mibObjectsView) makeAPIIndex() []api.ObjectIndex {
	var objects []api.ObjectIndex

	view.mib.Walk(func(id mibs.ID) {
		if object := view.mib.Object(id); object != nil {
			objects = append(objects, objectView{object}.makeAPIIndex())
		}
	})

	return objects
}

type hostObjectView struct {
	host   *Host
	object *mibs.Object
}

func (view hostObjectView) query() api.Object {
	var object = api.Object{
		ObjectIndex: objectView{view.object}.makeAPIIndex(),
	}

	if value, err := view.host.getObject(view.object); err != nil {
		object.Error = &api.Error{err}
	} else {
		object.Value = value
	}

	return object
}

func (view hostObjectView) GetREST() (web.Resource, error) {
	return view.query(), nil
}

type hostObjectsView struct {
	host *Host
}

func (view hostObjectsView) GetREST() (web.Resource, error) {
	var apiObjects []api.Object

	view.host.walkObjects(func(object *mibs.Object) {
		apiObject := hostObjectView{view.host, object}.query()

		apiObjects = append(apiObjects, apiObject)
	})

	return apiObjects, nil
}
