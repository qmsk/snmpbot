package main

import (
	"flag"
	"fmt"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/server"
)

type Options struct {
	cmd.Options

	Server server.Options
	Web    web.Options
}

func (options *Options) InitServerFlags() {
	flag.StringVar(&options.Server.ConfigFile, "config", "", "Load TOML config")

	flag.StringVar(&options.Web.Listen, "http-listen", ":8286", "HTTP server listen: [HOST]:PORT")
	flag.StringVar(&options.Web.Static, "http-static", "", "HTTP sever /static path: PATH")
}

var options Options

func init() {
	options.InitFlags()
	options.InitServerFlags()
}

func run(engine *server.Engine) error {
	// XXX: this is not a good API
	options.Web.Server(
		options.Web.RouteAPI("/api/", engine.WebAPI()),
		options.Web.RouteStatic("/"),
	)

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
