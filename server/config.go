package server

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/qmsk/snmpbot/client"
	"strings"
)

type ConfigKeysError struct {
	Source string
	Keys   []toml.Key
}

func (err ConfigKeysError) String() string {
	var strs = make([]string, len(err.Keys))

	for i, key := range err.Keys {
		strs[i] = key.String()
	}

	return strings.Join(strs, " ")
}

func (err ConfigKeysError) Error() string {
	return fmt.Sprintf("Unexpected keys in %s: %s", err.Source, err.String())
}

type Config struct {
	SNMP  client.Options
	Hosts map[string]HostConfig
}

func (config *Config) LoadTOML(path string) error {
	log.Infof("Load config from %v", path)

	if tomlMeta, err := toml.DecodeFile(path, config); err != nil {
		return err
	} else if undecodedKeys := tomlMeta.Undecoded(); len(undecodedKeys) > 0 {
		return ConfigKeysError{path, undecodedKeys}
	}

	for hostID, hostConfig := range config.Hosts {
		if hostConfig.SNMP == nil {
			hostConfig.SNMP = &config.SNMP
		}

		config.Hosts[hostID] = hostConfig
	}

	return nil
}
