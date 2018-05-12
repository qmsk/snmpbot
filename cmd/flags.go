package cmd

import (
	"flag"
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
	"log"
	"os"
)

type Options struct {
	Logging       logging.Options
	MIBs          mibs.Options
	MIBsLogging   logging.Options
	Client        client.Options
	ClientLogging logging.Options
}

func (options *Options) InitFlags() {
	options.MIBsLogging = logging.Options{
		Module:   "mibs",
		Defaults: &options.Logging,
	}
	options.ClientLogging = logging.Options{
		Module:   "client",
		Defaults: &options.Logging,
	}

	options.Logging.InitFlags()
	options.MIBsLogging.InitFlags()
	options.ClientLogging.InitFlags()

	options.MIBs.InitFlags()
	options.Client.InitFlags()
}

func (options *Options) Parse() []string {
	flag.Parse()

	mibs.SetLogging(options.MIBsLogging.MakeLogging())
	client.SetLogging(options.ClientLogging.MakeLogging())

	return flag.Args()
}

func (options Options) ClientEngine() (*client.Engine, error) {
	return client.NewUDPEngine(options.Client.UDP)
}

func (options Options) ClientConfig(url string) (client.Config, error) {
	return client.ParseConfig(options.Client, url)
}

func (options Options) ParseClientIDs(engine *client.Engine, args []string) (*client.Client, []mibs.ID, error) {
	if len(args) < 1 {
		return nil, nil, fmt.Errorf("Usage: [options] <addr> <oid...>")
	}

	if clientConfig, err := options.ClientConfig(args[0]); err != nil {
		return nil, nil, fmt.Errorf("Invalid addr %v: %v", args[0], err)
	} else if ids, err := options.ResolveIDs(args[1:]); err != nil {
		return nil, nil, err
	} else if client, err := client.NewClient(engine, clientConfig); err != nil {
		return nil, nil, fmt.Errorf("NewClient: %v", err)
	} else {
		return client, ids, nil
	}
}

func (options Options) runEngine(engine *client.Engine) {
	if err := engine.Run(); err != nil {
		log.Fatalf("FATAL client:Engine.Run: %v", err)
	} else {
		log.Fatalf("FATAL client:Engine.Run: stopped")
	}
}

func (options Options) withEngine(engine *client.Engine, f func() error) error {
	go options.runEngine(engine)
	defer engine.Close()

	return f()
}

func (options Options) WithEngine(args []string, f func(*client.Engine) error) error {
	if engine, err := options.ClientEngine(); err != nil {
		return err
	} else {
		return options.withEngine(engine, func() error {
			return f(engine)
		})
	}
}

func (options Options) WithClientOIDs(args []string, f func(*client.Client, ...snmp.OID) error) error {
	if engine, err := options.ClientEngine(); err != nil {
		return err
	} else if client, ids, err := options.ParseClientIDs(engine, args); err != nil {
		return err
	} else {
		var oids = make([]snmp.OID, len(ids))

		for i, id := range ids {
			oids[i] = id.OID
		}

		return options.withEngine(engine, func() error {
			return f(client, oids...)
		})
	}
}

func (options Options) WithClientIDs(args []string, f func(mibs.Client, ...mibs.ID) error) error {
	if engine, err := options.ClientEngine(); err != nil {
		return err
	} else if snmpClient, ids, err := options.ParseClientIDs(engine, args); err != nil {
		return err
	} else {
		var client = mibs.MakeClient(snmpClient)

		return options.withEngine(engine, func() error {
			return f(client, ids...)
		})
	}
}

func (options Options) WithClientID(args []string, f func(mibs.Client, mibs.ID) error) error {
	if engine, err := options.ClientEngine(); err != nil {
		return err
	} else if snmpClient, ids, err := options.ParseClientIDs(engine, args); err != nil {
		return err
	} else {
		var client = mibs.MakeClient(snmpClient)

		return options.withEngine(engine, func() error {
			for _, id := range ids {
				if err := f(client, id); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func (options *Options) Main(f func(args []string) error) {
	args := options.Parse()

	if err := options.MIBs.LoadMIBs(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err := f(args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(0)
}
