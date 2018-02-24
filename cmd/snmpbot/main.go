package main

import (
	"flag"
	"fmt"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/server"
	"github.com/qmsk/snmpbot/util/logging"
)

type Options struct {
	cmd.Options

	Server        server.Options
	ServerLogging logging.Options
	Web           web.Options
}

func (options *Options) InitFlags() {
	options.ServerLogging = logging.Options{
		Module:   "server",
		Defaults: &options.Options.Logging,
	}
	options.Options.InitFlags()
	options.Server.InitFlags()
	options.ServerLogging.InitFlags()

	flag.StringVar(&options.Web.Listen, "http-listen", ":8286", "HTTP server listen: [HOST]:PORT")
	flag.StringVar(&options.Web.Static, "http-static", "", "HTTP sever /static path: PATH")
}

func (options *Options) Apply() {
	options.Server.SNMP = options.ClientConfig()

	server.ApplyLogging(options.ServerLogging)
}

var options Options

func init() {
	options.InitFlags()
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
		options.Apply()

		if engine, err := options.Server.Engine(); err != nil {
			return fmt.Errorf("Failed to load server: %v", err)
		} else {
			return run(engine)
		}
	})
}
