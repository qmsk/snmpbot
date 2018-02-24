package logging

import (
	"flag"
)

type Options struct {
	Debug   bool
	Verbose bool
	Quiet   bool
}

func (options *Options) InitFlags() {
	flag.BoolVar(&options.Debug, "debug", false, "Log debug")
	flag.BoolVar(&options.Verbose, "verbose", false, "Log info")
	flag.BoolVar(&options.Quiet, "quiet", false, "Do not log warnings")
}

func (options *Options) MakeLogging() Logging {
	var logging = Logging{}

	if options.Debug {
		logging.Debug = MakeLogger("DEBUG: ")
	}
	if options.Debug || options.Verbose {
		logging.Info = MakeLogger("INFO: ")
	}
	if !options.Quiet {
		logging.Warn = MakeLogger("WARN: ")
		logging.Error = MakeLogger("ERROR: ")
	}

	return logging
}
