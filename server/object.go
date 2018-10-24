package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
	"path"
	"strings"
)

type ObjectID string

func AllObjects() Objects {
	var objects = make(Objects)

	mibs.WalkObjects(func(object *mibs.Object) {
		if object.NotAccessible {
			return
		}
		objects.add(object)
	})

	return objects
}

func MakeObjects(args ...*mibs.Object) Objects {
	var objects = make(Objects, len(args))

	for _, object := range args {
		objects.add(object)
	}

	return objects
}

type Objects map[ObjectID]*mibs.Object

func (objects Objects) add(object *mibs.Object) {
	objects[ObjectID(object.Key())] = object
}

func (objects Objects) exists(object *mibs.Object) bool {
	if _, exists := objects[ObjectID(object.Key())]; exists {
		return true
	} else {
		return false
	}
}

func (objects Objects) String() string {
	var ss = make([]string, 0, len(objects))

	for _, object := range objects {
		ss = append(ss, object.String())
	}

	return "{" + strings.Join(ss, ", ") + "}"
}

func (objects Objects) List() []*mibs.Object {
	var list = make([]*mibs.Object, 0, len(objects))

	for _, object := range objects {
		list = append(list, object)
	}

	return list
}

func (objects Objects) Filter(filters ...string) Objects {
	var filtered = make(Objects)

	for objectID, object := range objects {
		var match = false
		var name = object.String()

		for _, filter := range filters {
			if matched, _ := path.Match(filter, name); matched {
				match = true
			}
		}

		if match {
			filtered[objectID] = object
		}
	}

	return filtered
}

// Select objects belonging to tables
func (objects Objects) FilterTables(tables Tables) Objects {
	var filtered = make(Objects)

	for _, table := range tables {
		for _, object := range table.EntrySyntax {
			if object.NotAccessible {
				continue
			}

			if !objects.exists(object) {
				continue
			}

			filtered.add(object)
		}
	}

	log.Debugf("Filter %d => %d objects by tables: %#v", len(objects), len(filtered), tables)

	return filtered
}

type objectsRoute struct {
	engine Engine
}

func (route objectsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return &objectsHandler{
			engine:  route.engine,
			hosts:   route.engine.Hosts(),
			objects: route.engine.Objects(),
			tables:  route.engine.Tables(),
		}, nil
	} else if object, err := mibs.ResolveObject(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return &objectHandler{
			engine: route.engine,
			hosts:  route.engine.Hosts(),
			object: object,
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
	object *mibs.Object
}

func (view objectView) makeIndexKeys() []string {
	if view.object.IndexSyntax == nil {
		return nil
	}

	var keys = make([]string, len(view.object.IndexSyntax))

	for i, indexObject := range view.object.IndexSyntax {
		keys[i] = indexObject.String()
	}

	return keys
}

func (view objectView) makeAPIIndex() api.ObjectIndex {
	var index = api.ObjectIndex{
		ID:        view.object.String(),
		IndexKeys: view.makeIndexKeys(),
	}

	return index
}

func (view objectView) makeObjectIndex(indexValues mibs.IndexValues) api.ObjectIndexMap {
	if indexValues == nil {
		return nil
	}
	var indexMap = make(api.ObjectIndexMap)

	for i, indexObject := range view.object.IndexSyntax {
		indexMap[indexObject.String()] = indexValues[i]
	}

	return indexMap
}

func (view objectView) instanceFromResult(result ObjectResult) api.ObjectInstance {
	var object = api.ObjectInstance{
		HostID: string(result.Host.id),
		Value:  result.Value,
	}

	// XXX: should always match...?
	if result.Object == view.object {
		object.Index = view.makeObjectIndex(result.IndexValues)
	}

	return object
}

func (view objectView) errorFromResult(result ObjectResult) api.ObjectError {
	var ret = api.ObjectError{
		HostID: string(result.Host.id),
		Value:  result.Value,
	}

	// XXX: should always match...?
	if result.Object == view.object {
		ret.Index = view.makeObjectIndex(result.IndexValues)
	}

	ret.Error = api.Error{result.Error}

	return ret
}

type objectsView struct {
	objects Objects
}

func (view objectsView) makeAPIIndex() []api.ObjectIndex {
	var objects []api.ObjectIndex

	for _, object := range view.objects {
		objects = append(objects, objectView{object}.makeAPIIndex())
	}

	return objects
}

type objectHandler struct {
	engine Engine
	hosts  Hosts
	object *mibs.Object
	params api.ObjectQuery
}

func (handler *objectHandler) query() api.Object {
	var object = api.Object{
		ObjectIndex: objectView{handler.object}.makeAPIIndex(),
		Instances:   []api.ObjectInstance{},
	}

	for result := range handler.engine.QueryObjects(ObjectQuery{
		Hosts:   handler.hosts,
		Objects: MakeObjects(handler.object),
	}) {
		if result.Error != nil {
			object.Errors = append(object.Errors, objectView{result.Object}.errorFromResult(result))
		} else {
			object.Instances = append(object.Instances, objectView{result.Object}.instanceFromResult(result))
		}
	}

	return object
}

func (handler *objectHandler) QueryREST() interface{} {
	return &handler.params
}

func (handler *objectHandler) GetREST() (web.Resource, error) {
	log.Debugf("GET .../objects/%v %#v", handler.object, handler.params)

	if handler.params.Hosts != nil {
		handler.hosts = handler.hosts.Filter(handler.params.Hosts...)
	}

	return handler.query(), nil
}

type objectsHandler struct {
	engine  Engine
	hosts   Hosts
	objects Objects
	tables  Tables
	params  api.ObjectsQuery
}

func (handler *objectsHandler) query() ([]*api.Object, error) {
	var objectMap = make(map[ObjectID]*api.Object, len(handler.objects))
	var objects = make([]*api.Object, 0, len(handler.objects))
	var err error

	for objectID, o := range handler.objects {
		var object = api.Object{
			ObjectIndex: objectView{o}.makeAPIIndex(),
			Instances:   []api.ObjectInstance{},
		}

		objectMap[objectID] = &object
		objects = append(objects, &object)
	}

	for result := range handler.engine.QueryObjects(ObjectQuery{
		Hosts:   handler.hosts,
		Objects: handler.objects,
	}) {
		var object = objectMap[ObjectID(result.Object.Key())]

		if result.Error != nil {
			object.Errors = append(object.Errors, objectView{result.Object}.errorFromResult(result))
		} else {
			object.Instances = append(object.Instances, objectView{result.Object}.instanceFromResult(result))
		}
	}

	return objects, err
}

func (handler *objectsHandler) QueryREST() interface{} {
	return &handler.params
}

func (handler *objectsHandler) GetREST() (web.Resource, error) {
	log.Debugf("GET .../objects/ %#v", handler.params)

	if handler.params.Hosts != nil {
		handler.hosts = handler.hosts.Filter(handler.params.Hosts...)
	}

	if handler.params.Tables != nil {
		handler.objects = handler.objects.FilterTables(handler.tables.Filter(handler.params.Tables...))
	}

	if handler.params.Objects != nil {
		handler.objects = handler.objects.Filter(handler.params.Objects...)
	}

	return handler.query()
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
