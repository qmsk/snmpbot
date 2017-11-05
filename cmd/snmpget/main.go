package main

import (
	"fmt"
	client "github.com/qmsk/snmpbot/client"
	cmd "github.com/qmsk/snmpbot/cmd"
	snmp "github.com/qmsk/snmpbot/snmp_new"
	"log"
	"os"
)

type Options struct {
	cmd.Options
}

var options Options

func init() {
	options.InitFlags()
}

func snmpget(client *client.Client, args ...string) error {
	var oids = make([]snmp.OID, len(args))

	for i, arg := range args {
		if oid, err := snmp.ParseOID(arg); err != nil {
			return fmt.Errorf("Invalid OID %v: %v", arg, err)
		} else {
			oids[i] = oid
		}
	}

	if varBinds, err := client.Get(oids...); err != nil {
		return fmt.Errorf("client.Get: %v", err)
	} else {
		for _, varBind := range varBinds {
			if value, err := varBind.Value(); err != nil {
				log.Printf("VarBind[%v].Value: %v", varBind.OID, err)
			} else {
				fmt.Printf("%v = <%T> %v\n", varBind.OID(), value, value)
			}
		}
	}

	return nil
}

func run(options Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Usage: [options] <addr> <oid...>")
	}

	var clientConfig = options.ClientConfig()

	if err := clientConfig.Parse(args[0]); err != nil {
		return fmt.Errorf("Invalid addr %v: %v", args[0], err)
	}

	if client, err := clientConfig.Client(); err != nil {
		return fmt.Errorf("Client: %v", err)
	} else {
		go client.Run()
		defer client.Close()

		return snmpget(client, args[1:]...)
	}
}

func main() {
	args := options.Parse()

	if err := run(options, args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(0)
}
