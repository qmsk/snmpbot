package cmd

import (
	"flag"
	client "github.com/qmsk/snmpbot/client"
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
