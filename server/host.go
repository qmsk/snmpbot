package server

import (
	"fmt"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"log"
)

type HostConfig struct {
	Host string

	// optional, defaults to global config
	SNMP *client.Config
}

func newHost(id HostID) *Host {
	return &Host{id: id}
}

type Host struct {
	id         HostID
	config     HostConfig
	probedMIBs []*mibs.MIB
	snmpClient *client.Client
}

func (host *Host) String() string {
	return fmt.Sprintf("%v", host.id)
}

func (host *Host) init(config HostConfig) error {
	var snmpConfig = *config.SNMP

	if config.Host == "" {
		config.Host = string(host.id)
	}

	if err := snmpConfig.Parse(config.Host); err != nil {
		return err
	}

	host.config = config
	host.config.SNMP = &snmpConfig

	log.Printf("Host<%v>: Config SNMP: %#v", host, host.config.SNMP)

	if snmpClient, err := snmpConfig.Client(); err != nil {
		return fmt.Errorf("SNMP client for %v: %v", host, err)
	} else {
		log.Printf("Host<%v>: Connected client: %v", host, snmpClient)

		host.snmpClient = snmpClient
	}

	return nil
}

func (host *Host) makeMIBIDs() []mibs.ID {
	var ids []mibs.ID

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		ids = append(ids, mib.ID)
	})

	return ids
}

func (host *Host) probe() error {
	var client = mibs.Client{host.snmpClient}
	var ids = host.makeMIBIDs()

	log.Printf("Host<%v>: Probing MIBs: %v", host, ids)

	if probed, err := client.ProbeMany(ids); err != nil {
		return err
	} else {
		for _, id := range ids {
			log.Printf("Host<%v>: Probed %v = %v", host, id, probed[id.Key()])

			if probed[id.Key()] {
				host.probedMIBs = append(host.probedMIBs, id.MIB)
			}
		}
	}

	return nil
}

func (host *Host) start() {
	log.Printf("Host<%v>: Starting...", host)

	go host.run()
}

func (host *Host) run() {
	log.Printf("Host<%v>: Running...", host)

	if err := host.snmpClient.Run(); err != nil {
		// XXX: handle restarts?
		log.Printf("Host<%v>: SNMP client failed: %v", host, err)
	}

	log.Printf("Host<%v>: Stopped", host)
}

func (host *Host) stop() {
	log.Printf("Host<%v>: Stopping...", host)

	host.snmpClient.Close()
}

func (host *Host) walkObjects(f func(*mibs.Object)) {
	for _, mib := range host.probedMIBs {
		mib.Walk(func(id mibs.ID) {
			if object := mib.Object(id); object != nil {
				f(object)
			}
		})
	}
}

func (host *Host) walkTables(f func(*mibs.Table)) {
	for _, mib := range host.probedMIBs {
		mib.Walk(func(id mibs.ID) {
			if table := mib.Table(id); table != nil {
				f(table)
			}
		})
	}
}

func (host *Host) Objects() Objects {
	var objects = make(Objects)

	host.walkObjects(func(object *mibs.Object) {
		objects.add(object)
	})

	return objects
}

func (host *Host) Tables() Tables {
	var tables = make(Tables)

	host.walkTables(func(table *mibs.Table) {
		tables.add(table)
	})

	return tables
}

func (host *Host) resolveObject(name string) (*mibs.Object, error) {
	return mibs.ResolveObject(name)
}

func (host *Host) resolveTable(name string) (*mibs.Table, error) {
	return mibs.ResolveTable(name)
}

type hostRoute struct {
	engine *Engine
	host   *Host
}

func (route hostRoute) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return hostView{route.host}, nil
	case "objects":
		return hostObjectsRoute(route), nil
	case "tables":
		return hostTablesRoute(route), nil
	default:
		return nil, nil
	}
}

func (route hostRoute) GetREST() (web.Resource, error) {
	return hostView{route.host}.makeAPIIndex(), nil
}

type hostView struct {
	host *Host
}

func (view hostView) makeMIBs() []api.MIBIndex {
	var mibs = make([]api.MIBIndex, len(view.host.probedMIBs))

	for i, mib := range view.host.probedMIBs {
		mibs[i] = mibView{mib}.makeAPIIndex()
	}

	return mibs
}

func (view hostView) makeObjects() []api.ObjectIndex {
	var objects []api.ObjectIndex

	view.host.walkObjects(func(object *mibs.Object) {
		objects = append(objects, objectView{object}.makeAPIIndex())
	})

	return objects
}

func (view hostView) makeTables() []api.TableIndex {
	var tables []api.TableIndex

	view.host.walkTables(func(table *mibs.Table) {
		tables = append(tables, tableView{table}.makeAPIIndex())
	})

	return tables
}

func (view hostView) makeAPIIndex() api.HostIndex {
	return api.HostIndex{
		ID:   string(view.host.id),
		SNMP: view.host.snmpClient.String(),
		MIBs: view.makeMIBs(),
	}
}

func (view hostView) makeAPI() api.Host {
	return api.Host{
		HostIndex: view.makeAPIIndex(),
		Objects:   view.makeObjects(),
		Tables:    view.makeTables(),
	}
}

func (view hostView) GetREST() (web.Resource, error) {
	return view.makeAPI(), nil
}

type hostObjectsRoute hostRoute

func (route hostObjectsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return objectsHandler{
			engine:  route.engine,
			hosts:   MakeHosts(route.host),
			objects: route.host.Objects(),
		}, nil
	} else if object, err := route.host.resolveObject(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return objectHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			object: object,
		}, nil
	}
}

type hostTablesRoute hostRoute

func (route hostTablesRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return tablesHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			tables: route.host.Tables(),
		}, nil
	} else if table, err := route.host.resolveTable(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return tableHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			table:  table,
		}, nil
	}
}
