package logging

import (
	"flag"
)

type Options struct {
	Prefix  string
	Debug   bool
	Verbose bool
	Quiet   bool
}

func (options *Options) InitFlags(prefix string) {
	var flagPrefix = ""

	if prefix != "" {
		options.Prefix = prefix
		flagPrefix = prefix + "-"
	}

	flag.BoolVar(&options.Debug, flagPrefix+"debug", false, "Log debug")
	flag.BoolVar(&options.Verbose, flagPrefix+"verbose", false, "Log info")
	flag.BoolVar(&options.Quiet, flagPrefix+"quiet", false, "Do not log warnings")
}

func (options *Options) MakeLogging() Logging {
	var logging = Logging{}
	var logSuffix = ": "

	if options.Prefix != "" {
		logSuffix = " " + options.Prefix + logSuffix
	}

	if options.Debug {
		logging.Debug = MakeLogger("DEBUG" + logSuffix)
	}
	if options.Debug || options.Verbose {
		logging.Info = MakeLogger("INFO" + logSuffix)
	}
	if !options.Quiet {
		logging.Warn = MakeLogger("WARN" + logSuffix)
	}
	logging.Error = MakeLogger("ERROR" + logSuffix)

	return logging
}
