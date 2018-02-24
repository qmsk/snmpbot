package logging

import (
	"flag"
)

type Options struct {
	Module   string
	Defaults *Options

	Debug   bool
	Verbose bool
	Quiet   bool
}

func (options *Options) InitFlags() {
	var flagSuffix = ""
	var descSuffix = ""

	if options.Module != "" {
		flagSuffix = "." + options.Module
		descSuffix = " for " + options.Module
	}

	flag.BoolVar(&options.Debug, "debug"+flagSuffix, false, "Log debug"+descSuffix)
	flag.BoolVar(&options.Verbose, "verbose"+flagSuffix, false, "Log info"+descSuffix)
	flag.BoolVar(&options.Quiet, "quiet"+flagSuffix, false, "Do not log warnings"+descSuffix)
}

func (options *Options) applyDefaults() {
	if options.Defaults.Debug {
		options.Debug = true
	}
	if options.Defaults.Verbose {
		options.Verbose = true
	}
	if options.Defaults.Quiet {
		options.Quiet = true
	}
}

func (options *Options) MakeLogging() Logging {
	var logging = Logging{}
	var logSuffix = ": "

	if options.Module != "" {
		logSuffix = " " + options.Module + logSuffix
	}
	if options.Defaults != nil {
		options.applyDefaults()
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
