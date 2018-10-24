package server

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/client"
)

type Options struct {
	ConfigFile string
}

func (options *Options) InitFlags() {
	flag.StringVar(&options.ConfigFile, "config", "", "Load TOML config")
}

func (options Options) LoadConfig(clientOptions client.Options) (Config, error) {
	var config = Config{
		ClientOptions: clientOptions,
	}

	if options.ConfigFile == "" {
		log.Debugf("Not loading any config file")

	} else if err := config.LoadTOML(options.ConfigFile); err != nil {
		return config, fmt.Errorf("Failed to load config from %v: %v", options.ConfigFile, err)
	} else {
		log.Infof("Load config from %v", options.ConfigFile)
	}

	return config, nil
}

func (options Options) Engine(clientEngine *client.Engine, config Config) (Engine, error) {
	var engine = newEngine(clientEngine)

	if err := engine.loadConfig(config); err != nil {
		return nil, err
	}

	return engine, nil
}
