package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type objectsRoute struct {
}

func (route objectsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return objectsView{}, nil
	} else if object, err := mibs.ResolveObject(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return objectView{object}, nil
	}
}

func (route objectsRoute) makeIndex() api.IndexObjects {
	return api.IndexObjects{
		Objects: objectsView{}.makeAPIIndex(),
	}
}

func (route objectsRoute) GetREST() (web.Resource, error) {
	return route.makeIndex(), nil
}

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

func (view hostObjectsView) query() []api.Object {
	var apiObjects []api.Object

	view.host.walkObjects(func(object *mibs.Object) {
		apiObject := hostObjectView{view.host, object}.query()

		apiObjects = append(apiObjects, apiObject)
	})

	return apiObjects
}

func (view hostObjectsView) GetREST() (web.Resource, error) {
	return view.query(), nil
}
