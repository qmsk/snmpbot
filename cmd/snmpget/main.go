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

func snmpget(client *client.Client, oids ...snmp.OID) error {
	if varBinds, err := client.Get(oids...); err != nil {
		return fmt.Errorf("client.Get: %v", err)
	} else {
		for _, varBind := range varBinds {
			if value, err := varBind.Value(); err != nil {
				log.Printf("VarBind[%v].Value: %v", varBind.OID(), err)
			} else {
				fmt.Printf("%v = <%T> %v\n", cmd.FormatOID(varBind.OID()), value, value)
			}
		}
	}

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientOIDs(args, snmpget)
	})
}
