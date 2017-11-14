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

func snmpobject(client *mibs.Client, id mibs.ID) error {
	var object = id.Object()

	if object == nil {
		return fmt.Errorf("Not an object: %v", id)
	}

	if value, err := client.GetObject(object); err != nil {
		return err
	} else {
		fmt.Printf("%v = %v\n", object, value)
	}

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientID(args, snmpobject)
	})
}
