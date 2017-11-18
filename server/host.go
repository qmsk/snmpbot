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

func newHost(id hostID) *Host {
	return &Host{id: id}
}

type Host struct {
	id         hostID
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

func (host *Host) resolveObject(name string) (*mibs.Object, error) {
	return mibs.ResolveObject(name)
}

func (host *Host) resolveTable(name string) (*mibs.Table, error) {
	return mibs.ResolveTable(name)
}

func (host *Host) getObject(object *mibs.Object) (mibs.Value, error) {
	return mibs.Client{host.snmpClient}.GetObject(object)
}

func (host *Host) walkTable(table *mibs.Table, f func(mibs.IndexMap, mibs.EntryMap) error) error {
	return mibs.Client{host.snmpClient}.WalkTable(table, f)
}

type hostRoute struct {
	host *Host
}

func (route hostRoute) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return hostView{route.host}, nil
	case "objects":
		return hostObjectsRoute{route.host}, nil
	case "tables":
		return hostTablesRoute{route.host}, nil
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

func (view hostView) makeMIBs() []string {
	var mibs = make([]string, len(view.host.probedMIBs))

	for i, mib := range view.host.probedMIBs {
		mibs[i] = mib.String()
	}

	return mibs
}

func (view hostView) makeAPIIndex() api.HostIndex {
	return api.HostIndex{
		ID:   string(view.host.id),
		SNMP: view.host.snmpClient.String(),
		MIBs: view.makeMIBs(),
	}
}

func (view hostView) makeAPI() api.Host {
	var host = api.Host{
		HostIndex: view.makeAPIIndex(),
	}

	// TODO: objects, tables

	return host
}
