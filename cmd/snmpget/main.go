package main

import (
	"fmt"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/snmp"
)

type Options struct {
	cmd.Options
}

var options Options

func init() {
	options.InitFlags()
}

func snmpget(client *client.Client, oids ...snmp.OID) error {
	if varBinds, err := client.Get(oids...); err != nil {
		return fmt.Errorf("client.Get: %v", err)
	} else {
		for _, varBind := range varBinds {
			options.PrintVarBind(varBind)
		}
	}

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientOIDs(args, snmpget)
	})
}
