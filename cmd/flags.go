package cmd

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/client"
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

func (options Options) ParseClientOIDs(args []string) (*client.Client, []snmp.OID, error) {
	if len(args) < 1 {
		return nil, nil, fmt.Errorf("Usage: [options] <addr> <oid...>")
	}

	var clientConfig = options.ClientConfig()
	var oids = make([]snmp.OID, len(args)-1)

	if err := clientConfig.Parse(args[0]); err != nil {
		return nil, nil, fmt.Errorf("Invalid addr %v: %v", args[0], err)
	}

	for i, arg := range args[1:] {
		if oid, err := ParseOID(arg); err != nil {
			return nil, nil, fmt.Errorf("Invalid OID %v: %v", arg, err)
		} else {
			oids[i] = oid
		}
	}

	if client, err := clientConfig.Client(); err != nil {
		return nil, nil, fmt.Errorf("Client: %v", err)
	} else {
		return client, oids, nil
	}
}

func (options Options) WithClientOIDs(args []string, f func(*client.Client, ...snmp.OID) error) error {
	if client, oids, err := options.ParseClientOIDs(args); err != nil {
		return err
	} else {
		go client.Run()
		defer client.Close()

		return f(client, oids...)
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
