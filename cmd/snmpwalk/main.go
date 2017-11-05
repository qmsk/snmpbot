package main

import (
	"fmt"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/snmp"
	"log"
)

type Options struct {
	cmd.Options
}

var options Options

func init() {
	options.InitFlags()
}

func snmpwalk(client *client.Client, oids ...snmp.OID) error {
	return client.Walk(func(varBinds ...snmp.VarBind) error {
		for _, varBind := range varBinds {
			if value, err := varBind.Value(); err != nil {
				log.Printf("VarBind[%v].Value: %v", varBind.OID(), err)
			} else {
				fmt.Printf("%v = <%T> %v\n", varBind.OID(), value, value)
			}
		}

		return nil
	}, oids...)
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientOIDs(args, snmpwalk)
	})
}
