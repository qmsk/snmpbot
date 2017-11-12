package cmd

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/qmsk/snmpbot/snmp"
	"log"
	"os"
)

type Options struct {
	LogDebug, LogVerbose, LogQuiet bool

	SNMP client.Config
}

func (options *Options) InitFlags() {
	flag.BoolVar(&options.LogDebug, "debug", false, "Log debug")
	flag.BoolVar(&options.LogVerbose, "verbose", false, "Log info")
	flag.BoolVar(&options.LogQuiet, "quiet", false, "Do not log warnings")

	flag.StringVar(&options.SNMP.Community, "snmp-community", "public", "Default SNMP community")
	flag.DurationVar(&options.SNMP.Timeout, "snmp-timeout", client.DefaultTimeout, "SNMP request timeout")
	flag.IntVar(&options.SNMP.Retry, "snmp-retry", 0, "SNMP request retry")
}

func (options *Options) Parse() []string {
	flag.Parse()
	return flag.Args()
}

func (options Options) ClientConfig() client.Config {
	var config = options.SNMP

	if options.LogDebug {
		config.Logging.Debug = makeLogger("DEBUG: ")
	}
	if options.LogDebug || options.LogVerbose {
		config.Logging.Info = makeLogger("INFO: ")
	}
	if !options.LogQuiet {
		config.Logging.Warn = makeLogger("WARN: ")
		config.Logging.Error = makeLogger("ERROR: ")
	}

	return config
}

func (options Options) ParseClientIDs(args []string) (*client.Client, []mibs.ID, error) {
	if len(args) < 1 {
		return nil, nil, fmt.Errorf("Usage: [options] <addr> <oid...>")
	}

	var clientConfig = options.ClientConfig()

	if err := clientConfig.Parse(args[0]); err != nil {
		return nil, nil, fmt.Errorf("Invalid addr %v: %v", args[0], err)
	}

	if ids, err := options.ResolveIDs(args[1:]); err != nil {
		return nil, nil, err
	} else if client, err := clientConfig.Client(); err != nil {
		return nil, nil, fmt.Errorf("Client: %v", err)
	} else {
		return client, ids, nil
	}
}

func (options Options) WithClientOIDs(args []string, f func(*client.Client, ...snmp.OID) error) error {
	if client, ids, err := options.ParseClientIDs(args); err != nil {
		return err
	} else {
		var oids = make([]snmp.OID, len(ids))

		for i, id := range ids {
			oids[i] = id.OID
		}

		go client.Run()
		defer client.Close()

		return f(client, oids...)
	}
}

func (options Options) WithClientIDs(args []string, f func(*mibs.Client, ...mibs.ID) error) error {
	if snmpClient, ids, err := options.ParseClientIDs(args); err != nil {
		return err
	} else {
		go snmpClient.Run()
		defer snmpClient.Close()

		var client = &mibs.Client{snmpClient}

		return f(client, ids...)
	}
}

func (options Options) WithClientID(args []string, f func(*mibs.Client, mibs.ID) error) error {
	if snmpClient, ids, err := options.ParseClientIDs(args); err != nil {
		return err
	} else {
		go snmpClient.Run()
		defer snmpClient.Close()

		var client = &mibs.Client{snmpClient}

		for _, id := range ids {
			if err := f(client, id); err != nil {
				return err
			}
		}

		return nil
	}
}

func (options Options) Main(f func(args []string) error) {
	args := options.Parse()

	if err := f(args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(0)
}
