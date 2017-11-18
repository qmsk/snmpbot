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
	probedMIBs []mibWrapper
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

func (host *Host) probe(mibs mibsWrapper) error {
	return mibs.probeHost(host.snmpClient, func(mib mibWrapper) {
		log.Printf("Host<%v>: Probed MIB: %v", host, mib)
		host.probedMIBs = append(host.probedMIBs, mib)
	})
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

func (host *Host) resolveObject(name string) (*mibs.Object, error) {
	return mibs.ResolveObject(name)
}

func (host *Host) getObject(object *mibs.Object) (mibs.Value, error) {
	return mibs.Client{host.snmpClient}.GetObject(object)
}

func (host *Host) makeAPIProbedMIBs() []string {
	var probedMIBs = make([]string, len(host.probedMIBs))

	for i, mib := range host.probedMIBs {
		probedMIBs[i] = mib.String()
	}

	return probedMIBs
}

func (host *Host) makeAPIIndex() api.HostIndex {
	return api.HostIndex{
		ID:         string(host.id),
		SNMP:       host.snmpClient.String(),
		ProbedMIBs: host.makeAPIProbedMIBs(),
	}
}

func (host *Host) makeAPI() api.Host {
	return api.Host{
		HostIndex: host.makeAPIIndex(),
	}
}

func (host *Host) GetREST() (web.Resource, error) {
	return host.makeAPIIndex(), nil
}

func (host *Host) Index(name string) (web.Resource, error) {
	switch name {
	case "objects":
		return hostObjectsView{host}, nil
	default:
		return nil, nil
	}
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
		return objectView{view.host, object}, nil
	}
}

func (view hostObjectsView) GetREST() (web.Resource, error) {
	var ret = api.HostObjects{
		HostID: string(view.host.id),
	}

	view.host.walkObjects(func(object *mibs.Object) {
		if value, err := view.host.getObject(object); err != nil {
			ret.Objects = append(ret.Objects, objectWrapper{object}.makeAPIError(err))
		} else {
			ret.Objects = append(ret.Objects, objectWrapper{object}.makeAPI(value))
		}
	})

	return ret, nil
}

type objectView struct {
	host   *Host
	object *mibs.Object
}

func (view objectView) GetREST() (web.Resource, error) {
	if value, err := view.host.getObject(view.object); err != nil {
		return objectWrapper{view.object}.makeAPIError(err), nil
	} else {
		return objectWrapper{view.object}.makeAPI(value), nil
	}
}
