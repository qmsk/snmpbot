package server

import (
	"fmt"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/client"
	"log"
)

type HostID string

type hosts map[HostID]*Host

func (hosts hosts) makeAPIIndex() []api.HostIndex {
	var items = make([]api.HostIndex, 0, len(hosts))

	for _, host := range hosts {
		items = append(items, host.makeAPIIndex())
	}

	return items
}

type HostConfig struct {
	Host string

	// optional, defaults to global config
	SNMP *client.Config
}

func newHost(id HostID) *Host {
	return &Host{id: id}
}

type Host struct {
	id         HostID
	config     HostConfig
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

	log.Printf("Host<%v>: Config SNMP: %v", host, host.config.SNMP)

	if snmpClient, err := snmpConfig.Client(); err != nil {
		return fmt.Errorf("SNMP client for %v: %v", host, err)
	} else {
		log.Printf("Host<%v>: Connected client: %v", host, snmpClient)

		host.snmpClient = snmpClient
	}

	return nil
}

func (host *Host) makeAPIIndex() api.HostIndex {
	return api.HostIndex{
		ID:   string(host.id),
		SNMP: host.snmpClient.String(),
	}
}
