package server

import (
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
)

type HostConfig struct {
	Host string

	// optional metadata
	Location string

	// optional, defaults to global config
	SNMP *client.Options
}

func makeHost(id HostID) Host {
	host := Host{id: id}
	host.log = logging.WithPrefix(log, fmt.Sprintf("Host<%v>", id))

	return host
}

func newHost(engine *Engine, id HostID, config HostConfig) (*Host, error) {
	var host = makeHost(id)

	if err := host.init(engine, config); err != nil {
		return nil, err
	} else {
		return &host, nil
	}
}

type HostState struct {
	Online   bool
	Location string
}

type Host struct {
	id         HostID
	config     HostConfig
	log        logging.PrefixLogging
	snmpClient *client.Client

	probedMIBs []*mibs.MIB
	state      HostState
	started    bool
}

func (host *Host) String() string {
	return fmt.Sprintf("%v", host.id)
}

func (host *Host) init(engine *Engine, config HostConfig) error {
	var clientOptions = engine.clientDefaults

	if config.SNMP != nil {
		clientOptions = *config.SNMP
	}

	if config.Host == "" {
		config.Host = string(host.id)
	}

	host.config = config
	host.state = HostState{
		Location: host.config.Location,
	}

	host.log.Infof("Config: %#v", host.config)

	if clientConfig, err := client.ParseConfig(clientOptions, config.Host); err != nil {
		return err
	} else if snmpClient, err := client.NewClient(engine.clientEngine, clientConfig); err != nil {
		return fmt.Errorf("NewClient %v: %v", host, err)
	} else {
		host.log.Infof("Connected client: %v", snmpClient)

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

	host.log.Infof("Probing MIBs: %v", ids)

	if probed, err := client.ProbeMany(ids); err != nil {
		return err
	} else {
		for _, id := range ids {
			host.log.Debugf("Probed %v = %v", id, probed[id.Key()])

			if probed[id.Key()] {
				host.probedMIBs = append(host.probedMIBs, id.MIB)
			}
		}

	}

	// TODO: probe system::sysLocation?
	host.state.Online = true

	return nil
}

func (host *Host) IsUp() bool {
	return host.state.Online
}

func (host *Host) start() {
	host.log.Infof("Starting...")

	host.started = true

	// TODO: periodic re-probing in case host was offline when starting?
	go func() {
		if err := host.probe(); err != nil {
			host.log.Warnf("Failed to probe: %v", err)
		}
	}()
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
		if object.NotAccessible {
			return
		}

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

func (host *Host) getClient() (mibs.Client, error) {
	return mibs.Client{host.snmpClient}, nil
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
		ID:       string(view.host.id),
		SNMP:     view.host.snmpClient.String(),
		Online:   view.host.state.Online,
		Location: view.host.state.Location,
		MIBs:     view.makeMIBs(),
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
	if !view.host.started {
		// for dynamic-lookup hosts that have not yet been probed
		if err := view.host.probe(); err != nil {
			return nil, err
		}
	}
	return view.makeAPI(), nil
}

type hostObjectsRoute hostRoute

func (route hostObjectsRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return &objectsHandler{
			engine:  route.engine,
			hosts:   MakeHosts(route.host),
			objects: route.host.Objects(),
		}, nil
	} else if object, err := route.host.resolveObject(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return &objectHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			object: object,
		}, nil
	}
}

type hostTablesRoute hostRoute

func (route hostTablesRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return &tablesHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			tables: route.host.Tables(),
		}, nil
	} else if table, err := route.host.resolveTable(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return &tableHandler{
			engine: route.engine,
			hosts:  MakeHosts(route.host),
			table:  table,
		}, nil
	}
}
