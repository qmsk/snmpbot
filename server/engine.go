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
		objects[ObjectID(object.Key())] = object
	})

	return objects
}

func (engine *Engine) Query(query Query) <-chan Result {
	var resultChan = make(chan Result)

	go query.execute(resultChan)

	return resultChan
}
