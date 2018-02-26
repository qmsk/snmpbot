package server

import (
	"fmt"
	"github.com/qmsk/snmpbot/client"
	"sync"
)

func newEngine(clientEngine *client.Engine) *Engine {
	return &Engine{
		clientEngine: clientEngine,
		mibs:         AllMIBs(),
		hosts:        make(Hosts),
	}
}

type Engine struct {
	clientEngine  *client.Engine
	clientOptions client.Options

	mibs       MIBs
	hosts      Hosts
	hostsMutex sync.Mutex
}

func (engine *Engine) newHost(id HostID, config HostConfig) (*Host, error) {
	host, err := newHost(engine, id, config)
	if err != nil {
		return nil, err
	}

	host.start()

	return host, nil
}

func (engine *Engine) loadConfig(config Config) error {
	engine.clientOptions = config.ClientOptions

	for hostName, hostConfig := range config.Hosts {
		if host, err := engine.newHost(HostID(hostName), hostConfig); err != nil {
			return fmt.Errorf("Failed to load host %v: %v", hostName, err)
		} else {
			engine.hosts[host.id] = host
		}
	}

	return nil
}

func (engine *Engine) MIBs() MIBs {
	return engine.mibs
}

func (engine *Engine) AddHost(id HostID, config HostConfig) (*Host, error) {
	if host, err := engine.newHost(id, config); err != nil {
		return nil, err
	} else {
		engine.addHost(host)

		return host, nil
	}
}

func (engine *Engine) addHost(host *Host) {
	engine.hostsMutex.Lock()
	defer engine.hostsMutex.Unlock()

	engine.hosts[host.id] = host

}

func (engine *Engine) Hosts() Hosts {
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

func (engine *Engine) Objects() Objects {
	// TODO: limit by MIBs?
	return AllObjects()
}

func (engine *Engine) Tables() Tables {
	// TODO: limit by MIBs?
	return AllTables()
}

func (engine *Engine) QueryObjects(q ObjectQuery) <-chan ObjectResult {
	q.resultChan = make(chan ObjectResult)

	go q.query()

	return q.resultChan
}

func (engine *Engine) QueryTables(q TableQuery) <-chan TableResult {
	q.resultChan = make(chan TableResult)

	go q.query()

	return q.resultChan
}
