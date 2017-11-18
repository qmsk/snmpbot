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

func (view hostObjectsView) Index(name string) (web.Resource, error) {
	if name == "" {
		return view, nil
	} else if object, err := view.host.resolveObject(name); err != nil {
		return nil, web.Errorf(400, "%v", err)
	} else {
		return hostObjectView{view.host, object}, nil
	}
}

func (view hostObjectsView) GetREST() (web.Resource, error) {
	var apiObjects []api.Object

	view.host.walkObjects(func(object *mibs.Object) {
		apiObject := hostObjectView{view.host, object}.query()

		apiObjects = append(apiObjects, apiObject)
	})

	return apiObjects, nil
}
