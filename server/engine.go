package server

import (
	"fmt"
)

func newEngine() *Engine {
	return &Engine{
		hosts: make(hosts),
	}
}

type Engine struct {
	hosts       hosts
	mibRegistry mibRegistry
}

func (engine *Engine) init(config Config) error {
	for hostName, hostConfig := range config.Hosts {
		var host = newHost(hostID(hostName))

		if err := host.init(hostConfig); err != nil {
			return fmt.Errorf("Failed to load host %v: %v", hostName, err)
		}

		engine.hosts[host.id] = host
	}

	return nil
}
