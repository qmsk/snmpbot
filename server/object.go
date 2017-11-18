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

func (view objectView) makeAPI(value mibs.Value) api.Object {
	return api.Object{
		ObjectIndex: view.makeAPIIndex(),
		Value:       value,
	}
}

func (view objectView) makeAPIError(err error) api.Object {
	return api.Object{
		ObjectIndex: view.makeAPIIndex(),
		Error:       &api.Error{err},
	}
}

type hostObjectView struct {
	host   *Host
	object *mibs.Object
}

func (view hostObjectView) getAPI() api.Object {
	if value, err := view.host.getObject(view.object); err != nil {
		return objectView{view.object}.makeAPIError(err)
	} else {
		return objectView{view.object}.makeAPI(value)
	}
}

func (view hostObjectView) GetREST() (web.Resource, error) {
	return view.getAPI(), nil
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
	var apiObjects = api.HostObjects{
		HostID: string(view.host.id),
	}

	view.host.walkObjects(func(object *mibs.Object) {
		apiObject := hostObjectView{view.host, object}.getAPI()

		apiObjects.Objects = append(apiObjects.Objects, apiObject)
	})

	return apiObjects, nil
}
