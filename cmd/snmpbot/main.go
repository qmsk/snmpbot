package main

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/server"
)

type Options struct {
	cmd.Options

	Server server.Options
}

func (options *Options) InitServerFlags() {
	flag.StringVar(&options.Server.ConfigFile, "config", "", "Load TOML config")
}

var options Options

func init() {
	options.InitFlags()
	options.InitServerFlags()
}

func run(engine *server.Engine) error {
	return nil
}

func main() {
	options.Main(func(args []string) error {
		options.Server.SNMP = options.SNMP

		if engine, err := options.Server.Engine(); err != nil {
			return fmt.Errorf("Failed to load server: %v", err)
		} else {
			return run(engine)
		}
	})
}
