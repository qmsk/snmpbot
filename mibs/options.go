package mibs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	MIBPath string
}

func (options *Options) InitFlags() {
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
