package server

import (
	"fmt"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/util/logging"
)

type HostConfig struct {
	Host string

	// optiona metadata
	Location string

	// optional, defaults to global config
	SNMP *client.Config
}

func newHost(id HostID) *Host {
	host := Host{id: id}
	host.log = logging.WithPrefix(log, fmt.Sprintf("Host<%v>", id))

	return &host
}

type HostState struct {
	Online   bool
	Location string
}

type Host struct {
	id         HostID
	config     HostConfig
	log        logging.PrefixLogging
	probedMIBs []*mibs.MIB
	snmpClient *client.Client
	state      HostState
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

	host.log.Infof("Config SNMP: %#v", host.config.SNMP)

	if snmpClient, err := snmpConfig.Client(); err != nil {
		return fmt.Errorf("SNMP client for %v: %v", host, err)
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

func (host *Host) probe() {
	var client = mibs.Client{host.snmpClient}
	var ids = host.makeMIBIDs()

	host.log.Infof("Probing MIBs: %v", ids)

	if probed, err := client.ProbeMany(ids); err != nil {
		host.log.Warnf("Failed to probe: %v", err)
	} else {
		for _, id := range ids {
			host.log.Debugf("Probed %v = %v", id, probed[id.Key()])

			if probed[id.Key()] {
				host.probedMIBs = append(host.probedMIBs, id.MIB)
			}
		}
	}

	host.state = HostState{
		Online: true,
		// TODO: probe system::sysLocation?
		Location: host.config.Location,
	}
}

func (host *Host) IsUp() bool {
	return host.state.Online
}

func (host *Host) start() {
	host.log.Infof("Starting...")

	go host.run()

	// TODO: period re-probing in case host was offline when starting?
	go host.probe()
}

func (host *Host) run() {
	host.log.Infof("Running...")

	if err := host.snmpClient.Run(); err != nil {
		// XXX: handle restarts?
		host.log.Errorf("SNMP client failed: %v", err)
	}

	host.log.Infof("Stopped")
}

func (host *Host) stop() {
	host.log.Infof("Stopping...")

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
