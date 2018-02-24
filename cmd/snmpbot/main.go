package main

import (
	"flag"
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/server"
)

type Options struct {
	cmd.Options

	Server        server.Options
	ServerLogging logging.Options
	Web           web.Options
	WebLogging    logging.Options
}

func (options *Options) InitFlags() {
	options.ServerLogging = logging.Options{
		Module:   "server",
		Defaults: &options.Options.Logging,
	}
	options.WebLogging = logging.Options{
		Module:   "web",
		Defaults: &options.Options.Logging,
	}
	options.Options.InitFlags()
	options.Server.InitFlags()
	options.ServerLogging.InitFlags()
	options.WebLogging.InitFlags()

	flag.StringVar(&options.Web.Listen, "http-listen", ":8286", "HTTP server listen: [HOST]:PORT")
	flag.StringVar(&options.Web.Static, "http-static", "", "HTTP sever /static path: PATH")
}

func (options *Options) Apply() {
	server.SetLogging(options.ServerLogging.MakeLogging())
	web.SetLogging(options.WebLogging.MakeLogging())
}

var options Options

func init() {
	options.InitFlags()
}

func run(serverEngine *server.Engine) error {
	// XXX: this is not a good API, it just returns immediately if there is no -http-listen?
	options.Web.Server(
		options.Web.RouteAPI("/api/", serverEngine.WebAPI()),
		options.Web.RouteStatic("/"),
	)

	return nil
}

func main() {
	options.Main(func(args []string) error {
		options.Apply()

		if clientEngine, err := options.ClientEngine(); err != nil {
			return fmt.Errorf("Failed to start client engine: %v", err)
		} else if config, err := options.Server.LoadConfig(options.Client); err != nil {
			return fmt.Errorf("Failed to load server config: %v", err)
		} else if serverEngine, err := options.Server.Engine(clientEngine, config); err != nil {
			return fmt.Errorf("Failed to load server: %v", err)
		} else {
			go clientEngine.Run()
			defer clientEngine.Close()

			return run(serverEngine)
		}
	})
}
