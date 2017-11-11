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

func snmptable(client *mibs.Client, id mibs.ID) error {
	table := id.Table()

	if table == nil {
		return fmt.Errorf("Not a table: %v", id)
	}

  for _, indexObject := range table.IndexSyntax {
    fmt.Printf("%v\t", indexObject)
  }
  for _, entryObject := range table.EntrySyntax {
    fmt.Printf("\t%v", entryObject)
  }
  fmt.Printf("\n")

	return client.WalkTable(table, func(indexMap mibs.IndexMap, entryMap mibs.EntryMap) error {
    for _, indexObject := range table.IndexSyntax {
      fmt.Printf("%v\t", indexMap[indexObject.ID.Key()])
    }
    for _, entryObject := range table.EntrySyntax {
      fmt.Printf("%v\t", entryMap[entryObject.ID.Key()])
    }
    fmt.Printf("\n")

		return nil
	})
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientID(args, snmptable)
	})
}
