package server

import (
	"fmt"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
)

func newEngine(clientEngine *client.Engine) *Engine {
	return &Engine{
		clientEngine: clientEngine,
		hosts:        make(Hosts),
	}
}

type Engine struct {
	clientEngine   *client.Engine
	clientDefaults client.Options

	hosts Hosts
}

func (engine *Engine) addHost(id HostID, config HostConfig) error {
	if host, err := newHost(engine, id, config); err != nil {
		return err
	} else {
		host.start()

		engine.hosts[id] = host
	}

	return nil
}

func (engine *Engine) init(config Config) error {
	engine.clientDefaults = config.SNMP

	for hostName, hostConfig := range config.Hosts {
		if err := engine.addHost(HostID(hostName), hostConfig); err != nil {
			return fmt.Errorf("Failed to load host %v: %v", hostName, err)
		}
	}

	return nil
}

func (engine *Engine) Hosts() Hosts {
	var hosts = make(Hosts)

	for hostID, host := range engine.hosts {
		if !host.IsUp() {
			continue
		}
		hosts[hostID] = host
	}

	return hosts
}

func (engine *Engine) Objects() Objects {
	var objects = make(Objects)

	mibs.WalkObjects(func(object *mibs.Object) {
		if object.NotAccessible {
			return
		}
		objects.add(object)
	})

	return objects
}

func (engine *Engine) Tables() Tables {
	var tables = make(Tables)

	mibs.WalkTables(func(table *mibs.Table) {
		tables.add(table)
	})

	return tables
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
