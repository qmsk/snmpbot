package server

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/client"
)

type Options struct {
	SNMP client.Config

	ConfigFile string
}

func (options *Options) InitFlags() {
	flag.StringVar(&options.ConfigFile, "config", "", "Load TOML config")
}

func (options Options) LoadConfig() (Config, error) {
	var config = Config{
		SNMP: options.SNMP,
	}

	if err := config.LoadTOML(options.ConfigFile); err != nil {
		return config, fmt.Errorf("Failed to load config from %v: %v", options.ConfigFile, err)
	}

	return config, nil
}

func (options Options) Engine() (*Engine, error) {
	var engine = newEngine()

	if config, err := options.LoadConfig(); err != nil {
		return nil, err
	} else if err := engine.init(config); err != nil {
		return nil, err
	}

	return engine, nil
}
