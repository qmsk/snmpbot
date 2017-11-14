package server

import (
	"fmt"
)

func newEngine() *Engine {
	return &Engine{
		hosts: make(map[HostID]*Host),
	}
}

type Engine struct {
	hosts map[HostID]*Host
}

func (engine *Engine) init(config Config) error {
	for hostName, hostConfig := range config.Hosts {
		var host = newHost(HostID(hostName))

		if err := host.init(hostConfig); err != nil {
			return fmt.Errorf("Failed to load host %v: %v", hostName, err)
		}
	}

	return nil
}
