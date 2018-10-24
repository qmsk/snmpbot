package server

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"sync"
)

type engineClient interface {
	String() string
	Probe(ids []mibs.ID) ([]bool, error)
	WalkObjects(objects []*mibs.Object, f func(*mibs.Object, mibs.IndexValues, mibs.Value, error) error) error
	WalkTable(table *mibs.Table, f func(mibs.IndexValues, mibs.EntryValues, error) error) error
}

type Engine interface {
	ClientOptions() client.Options
	client(config client.Config) (engineClient, error)

	MIBs() MIBs
	Objects() Objects
	Tables() Tables

	Hosts() Hosts
	AddHost(host *Host) bool
	SetHost(host *Host)
	DelHost(host *Host) bool

	QueryObjects(query ObjectQuery) <-chan ObjectResult
	QueryTables(query TableQuery) <-chan TableResult
}

func newEngine(clientEngine *client.Engine) *engine {
	return &engine{
		clientEngine: clientEngine,
		mibs:         AllMIBs(),
		hosts:        make(Hosts),
	}
}

type engine struct {
	clientEngine  *client.Engine
	clientOptions client.Options

	mibs       MIBs
	hosts      Hosts
	hostsMutex sync.Mutex
}

func (engine *engine) loadConfig(config Config) error {
	engine.clientOptions = config.ClientOptions

	for hostName, hostConfig := range config.Hosts {
		go engine.loadHost(HostID(hostName), hostConfig)
	}

	return nil
}

func (engine *engine) loadHost(id HostID, config HostConfig) {
	host, err := loadHost(engine, id, config)

	if err != nil {
		log.Warnf("Failed to load host %v: %v", id, err)

		host.err = err
	} else {
		log.Infof("Loaded host %v", id)
	}

	if !engine.AddHost(host) {
		log.Errorf("Duplicate host %v!", id)
	}
}

func (engine *engine) ClientOptions() client.Options {
	return engine.clientOptions
}

func (engine *engine) client(config client.Config) (engineClient, error) {
	if c, err := client.NewClient(engine.clientEngine, config); err != nil {
		return nil, err
	} else {
		return mibs.MakeClient(c), nil
	}
}

func (engine *engine) MIBs() MIBs {
	return engine.mibs
}

// Returns false if host already exists
func (engine *engine) AddHost(host *Host) bool {
	engine.hostsMutex.Lock()
	defer engine.hostsMutex.Unlock()

	if _, exists := engine.hosts[host.id]; !exists {
		engine.hosts[host.id] = host
		return true
	} else {
		return false
	}
}

func (engine *engine) SetHost(host *Host) {
	engine.hostsMutex.Lock()
	defer engine.hostsMutex.Unlock()

	engine.hosts[host.id] = host
}

// Returns false if host does not exist
func (engine *engine) DelHost(host *Host) bool {
	engine.hostsMutex.Lock()
	defer engine.hostsMutex.Unlock()

	if _, exists := engine.hosts[host.id]; exists {
		delete(engine.hosts, host.id)
		return true
	} else {
		return false
	}
}

func (engine *engine) Hosts() Hosts {
	var hosts = make(Hosts)

	engine.hostsMutex.Lock()
	defer engine.hostsMutex.Unlock()

	for hostID, host := range engine.hosts {
		if !host.IsUp() {
			continue
		}
		hosts[hostID] = host
	}

	return hosts
}

func (engine *engine) Objects() Objects {
	// TODO: limit by MIBs?
	return AllObjects()
}

func (engine *engine) Tables() Tables {
	// TODO: limit by MIBs?
	return AllTables()
}

func (engine *engine) QueryObjects(query ObjectQuery) <-chan ObjectResult {
	log.Infof("Query objects %v @ %v", query.Objects, query.Hosts)

	var q = objectQuery{
		ObjectQuery: query,
		resultChan:  make(chan ObjectResult),
	}

	go q.query()

	return q.resultChan
}

func (engine *engine) QueryTables(query TableQuery) <-chan TableResult {
	log.Infof("Query tables %v @ %v", query.Tables, query.Hosts)

	var q = tableQuery{
		TableQuery: query,
		resultChan: make(chan TableResult),
	}

	go q.query()

	return q.resultChan
}
