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

func snmpprobe(client mibs.Client, ids ...mibs.ID) error {
	if len(ids) == 0 {
		var mibList []*mibs.MIB

		mibs.WalkMIBs(func(mib *mibs.MIB) {
			mibList = append(mibList, mib)
		})

		if probed, err := client.ProbeMIBs(mibList); err != nil {
			return err
		} else {
			for i, ok := range probed {
				fmt.Printf("%v = %v\n", mibList[i], ok)
			}
		}

	} else {
		if probed, err := client.Probe(ids); err != nil {
			return err
		} else {
			for i, ok := range probed {
				fmt.Printf("%v = %v\n", ids[i], ok)
			}
		}
	}

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientIDs(args, snmpprobe)
	})
}
