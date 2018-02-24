package mibs

import (
	"flag"
	"fmt"
	"github.com/qmsk/snmpbot/util/logging"
	"os"
	"path/filepath"
)

type Options struct {
	Logging logging.Options
	MIBPath string
}

func (options *Options) InitFlags() {
	options.Logging.InitFlags("snmp-mibs")

	flag.StringVar(&options.MIBPath, "snmp-mibs", os.Getenv("SNMPBOT_MIBS"), "Load MIBs from PATH[:PATH[...]]")
}

func (options *Options) LoadMIBs() error {
	if options.MIBPath == "" {
		return fmt.Errorf("Must provide -snmp-mibs/$SNMPBOT_MIBS with path to .../snmpbot-mibs/*.json")
	}

	for _, path := range filepath.SplitList(options.MIBPath) {
		if err := Load(path); err != nil {
			return err
		}
	}

	return nil
}

func (options *Options) Apply() error {
	SetLogging(options.Logging.MakeLogging())

	if err := options.LoadMIBs(); err != nil {
		return err
	}

	return nil
}
