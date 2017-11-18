package server

import (
	"fmt"
	"github.com/qmsk/snmpbot/mibs"
)

func newEngine() *Engine {
	return &Engine{
		hosts: make(Hosts),
	}
}

type Engine struct {
	hosts Hosts
}

func (engine *Engine) init(config Config) error {
	for hostName, hostConfig := range config.Hosts {
		var host = newHost(HostID(hostName))

		if err := host.init(hostConfig); err != nil {
			return fmt.Errorf("Failed to load host %v: %v", hostName, err)
		}

		host.start()

		engine.hosts[host.id] = host

		// XXX: failures?
		if err := host.probe(); err != nil {
			return fmt.Errorf("Failed to probe host %v: %v", host, err)
		}
	}

	return nil
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
