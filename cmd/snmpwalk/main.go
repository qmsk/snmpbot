package main

import (
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

func snmpwalk(client *client.Client, oids ...snmp.OID) error {
	return client.WalkObjects(oids, func(varBinds []snmp.VarBind) error {
		for _, varBind := range varBinds {
			options.PrintVarBind(varBind)
		}

		return nil
	})
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientOIDs(args, snmpwalk)
	})
}
