package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type ObjectID string

func MakeObjects(args ...*mibs.Object) Objects {
	var objects = make(Objects, len(args))

	for _, object := range args {
		objects[ObjectID(object.Key())] = object
	}

	return objects
}

type Objects map[ObjectID]*mibs.Object

func (objects Objects) add(object *mibs.Object) {
	objects[ObjectID(object.Key())] = object
}

func (objects Objects) List() []*mibs.Object {
	var list = make([]*mibs.Object, 0, len(objects))

	for _, object := range objects {
		list = append(list, object)
	}

	return list
}

type objectsRoute struct {
	engine *Engine
}

func (route objectsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return objectsView{
			engine:  route.engine,
			hosts:   route.engine.hosts,
			objects: route.engine.Objects(),
		}, nil
	} else if object, err := mibs.ResolveObject(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		// XXX: should not return array for node
		return objectsView{
			engine:  route.engine,
			hosts:   route.engine.hosts,
			objects: MakeObjects(object),
		}, nil
	}
}

func (route objectsRoute) makeIndex() api.IndexObjects {
	return api.IndexObjects{
		Objects: objectsView{objects: route.engine.Objects()}.makeAPIIndex(),
	}
}

func (route objectsRoute) GetREST() (web.Resource, error) {
	return route.makeIndex(), nil
}

type objectView struct {
	*mibs.Object
}

func (view objectView) makeIndexKeys() []string {
	if view.Object.IndexSyntax == nil {
		return nil
	}

	var keys = make([]string, len(view.Object.IndexSyntax))

	for i, indexObject := range view.Object.IndexSyntax {
		keys[i] = indexObject.String()
	}

	return keys
}

func (view objectView) makeAPIIndex() api.ObjectIndex {
	var index = api.ObjectIndex{
		ID:        view.Object.String(),
		IndexKeys: view.makeIndexKeys(),
	}

	return index
}

func (view objectView) makeObjectIndex(indexValues mibs.IndexValues) api.ObjectIndexMap {
	if indexValues == nil {
		return nil
	}
	var indexMap = make(api.ObjectIndexMap)

	for i, indexObject := range view.Object.IndexSyntax {
		indexMap[indexObject.String()] = indexValues[i]
	}

	return indexMap
}

func (view objectView) fromResult(result Result) api.Object {
	var object = api.Object{
		HostID:      string(result.Host.id),
		ObjectIndex: view.makeAPIIndex(),
		Value:       result.Value,
	}

	if result.Object == view.Object {
		object.Index = view.makeObjectIndex(result.IndexValues)
	}

	if result.Error != nil {
		object.Error = &api.Error{result.Error}
	}

	return object
}

type objectsView struct {
	engine  *Engine
	hosts   Hosts
	objects Objects
}

func (view objectsView) makeAPIIndex() []api.ObjectIndex {
	var objects []api.ObjectIndex

	for _, object := range view.objects {
		objects = append(objects, objectView{object}.makeAPIIndex())
	}

	return objects
}

func (view objectsView) query() []api.Object {
	var objects = []api.Object{}

	for result := range view.engine.Query(Query{
		Hosts:   view.hosts,
		Objects: view.objects,
	}) {
		if result.Object != nil {
			objects = append(objects, objectView{result.Object}.fromResult(result))
		} else {
			// TODO: API errors?
		}
	}

	return objects
}

func (view objectsView) GetREST() (web.Resource, error) {
	return view.query(), nil
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
