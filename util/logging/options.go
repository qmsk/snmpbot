package logging

import (
	"flag"
)

type Options struct {
	Module string

	Debug   bool
	Verbose bool
	Quiet   bool
}

func (options *Options) InitFlags(module string) {
	var flagSuffix = ""
	var descSuffix = ""

	if module != "" {
		options.Module = module
		flagSuffix = "." + module
		descSuffix = " for " + module
	}

	flag.BoolVar(&options.Debug, "debug"+flagSuffix, false, "Log debug"+descSuffix)
	flag.BoolVar(&options.Verbose, "verbose"+flagSuffix, false, "Log info"+descSuffix)
	flag.BoolVar(&options.Quiet, "quiet"+flagSuffix, false, "Do not log warnings"+descSuffix)
}

func (options *Options) MakeLogging() Logging {
	var logging = Logging{}
	var logSuffix = ": "

	if options.Module != "" {
		logSuffix = " " + options.Module + logSuffix
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
