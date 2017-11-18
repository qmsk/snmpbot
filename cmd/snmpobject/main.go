package main

import (
	"fmt"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/mibs"
	"log"
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

	return client.WalkObjects(func(object *mibs.Object, indexValues mibs.IndexValues, value mibs.Value, err error) error {
		if err != nil {
			log.Printf("%v: %v", object, err)
		} else {
			fmt.Printf("%v%v = %v\n", object, object.IndexSyntax.FormatValues(indexValues), value)
		}

		return nil
	}, object)
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientID(args, snmpobject)
	})
}
