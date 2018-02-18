package main

import (
	"fmt"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/mibs"
)

type Options struct {
	cmd.Options
}

var options Options

func init() {
	options.InitFlags()
}

func snmpprobe(client *mibs.Client, ids ...mibs.ID) error {
	if len(ids) == 0 {
		mibs.WalkMIBs(func(mib *mibs.MIB) {
			ids = append(ids, mib.ID)
		})
	}

	for _, id := range ids {
		if ok, err := client.Probe(id); err != nil {
			return err
		} else {
			fmt.Printf("%v = %v\n", id, ok)
		}
	}

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientIDs(args, snmpprobe)
	})
}
