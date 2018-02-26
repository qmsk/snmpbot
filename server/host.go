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
	SNMP string

	// optional metadata
	Location string

	// optional, defaults to global config
	ClientOptions *client.Options // TODO: rename to ClientOptions
}

func newHost(id HostID) *Host {
	host := Host{id: id}
	host.log = logging.WithPrefix(log, fmt.Sprintf("Host<%v>", id))

	return &host
}

func loadHost(engine *Engine, id HostID, config HostConfig) (*Host, error) {
	var host = newHost(id)

	if err := host.init(engine, config); err != nil {
		return host, err
	} else if err := host.probe(engine.MIBs()); err != nil {
		return host, err
	} else {
		return host, nil
	}
}

type Host struct {
	id     HostID
	log    logging.PrefixLogging
	config HostConfig
	client *client.Client

	mibs   MIBs
	err    error
	online bool
}

func (host *Host) String() string {
	return fmt.Sprintf("%v", host.id)
}

func (host *Host) init(engine *Engine, config HostConfig) error {
	var clientOptions = engine.clientOptions

	if config.ClientOptions != nil {
		clientOptions = *config.ClientOptions
	}

	if config.SNMP == "" {
		config.SNMP = string(host.id)
	}

	host.config = config

	host.log.Infof("Config: %#v", host.config)

	if clientConfig, err := client.ParseConfig(clientOptions, config.SNMP); err != nil {
		return err
	} else if client, err := client.NewClient(engine.clientEngine, clientConfig); err != nil {
		return fmt.Errorf("NewClient %v: %v", host, err)
	} else {
		host.log.Infof("Connected client: %v", client)

		host.client = client
	}

	return nil
}

func (host *Host) probe(probeMIBs MIBs) error {
	var client = mibs.Client{host.client}

	host.log.Infof("Probing MIBs: %v", probeMIBs)

	if probed, err := client.ProbeMany(probeMIBs.ListIDs()); err != nil {
		return err
	} else {
		host.mibs = probeMIBs.FilterProbed(probed)
	}

	// TODO: probe system::sysLocation?
	host.online = true

	return nil
}

func (host *Host) IsUp() bool {
	return host.online
}

func (host *Host) MIBs() MIBs {
	return host.mibs
}

func (host *Host) Objects() Objects {
	return host.mibs.Objects()
}

func (host *Host) Tables() Tables {
	return host.mibs.Tables()
}

func (host *Host) resolveObject(name string) (*mibs.Object, error) {
	return mibs.ResolveObject(name)
}

func (host *Host) resolveTable(name string) (*mibs.Table, error) {
	return mibs.ResolveTable(name)
}

func (host *Host) getClient() (mibs.Client, error) {
	return mibs.Client{host.client}, nil
}

type hostRoute struct {
	engine *Engine
	host   *Host
}

func (route hostRoute) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return hostView{route.engine, route.host}, nil
	case "objects":
		return hostObjectsRoute(route), nil
	case "tables":
		return hostTablesRoute(route), nil
	default:
		return nil, nil
	}
}

func (route hostRoute) GetREST() (web.Resource, error) {
	return hostView{host: route.host}.makeAPIIndex(), nil
}

func (route hostRoute) DeleteREST() (web.Resource, error) {
	if exists := route.engine.DelHost(route.host); !exists {
		return nil, web.Errorf(404, "Host not configured: %v", route.host.id)
	}

	return nil, nil
}

type hostView struct {
	engine *Engine
	host   *Host
}

func (view hostView) makeMIBs() []api.MIBIndex {
	var mibs []api.MIBIndex

	for _, mib := range view.host.MIBs() {
		mibs = append(mibs, mibView{mib}.makeAPIIndex())
	}

	return mibs
}

func (view hostView) makeObjects() []api.ObjectIndex {
	var objects []api.ObjectIndex

	for _, object := range view.host.Objects() {
		objects = append(objects, objectView{object}.makeAPIIndex())
	}

	return objects
}

func (view hostView) makeTables() []api.TableIndex {
	var tables []api.TableIndex

	for _, table := range view.host.Tables() {
		tables = append(tables, tableView{table}.makeAPIIndex())
	}

	return tables
}

func (view hostView) makeAPIError() *api.Error {
	if view.host.err != nil {
		return &api.Error{view.host.err}
	} else {
		return nil
	}
}

func (view hostView) makeAPIIndex() api.HostIndex {
	return api.HostIndex{
		ID:       string(view.host.id),
		SNMP:     view.host.client.String(),
		Location: view.host.config.Location,
		Online:   view.host.online,
		Error:    view.makeAPIError(),
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
