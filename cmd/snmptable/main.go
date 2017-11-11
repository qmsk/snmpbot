package main

import (
	"fmt"
	"github.com/qmsk/snmpbot/cmd"
	"github.com/qmsk/snmpbot/mibs"
	"os"
	"text/tabwriter"
)

type Options struct {
	cmd.Options
}

var options Options

func init() {
	options.InitFlags()
}

func snmptable(client *mibs.Client, id mibs.ID) error {
	var table = id.Table()
	var writer = tabwriter.NewWriter(os.Stdout, 8, 4, 1, ' ', 0)

	if table == nil {
		return fmt.Errorf("Not a table: %v", id)
	}

	for _, indexObject := range table.IndexSyntax {
		fmt.Fprintf(writer, "%v\t", indexObject)
	}
	fmt.Fprintf(writer, "|")
	for _, entryObject := range table.EntrySyntax {
		fmt.Fprintf(writer, "\t%v", entryObject)
	}
	fmt.Fprintf(writer, "\n")

	for range table.IndexSyntax {
		fmt.Fprintf(writer, "---\t")
	}
	fmt.Fprintf(writer, "|")
	for range table.EntrySyntax {
		fmt.Fprintf(writer, "\t---")
	}
	fmt.Fprintf(writer, "\n")

	walkRow := func(indexMap mibs.IndexMap, entryMap mibs.EntryMap) error {
		for _, indexObject := range table.IndexSyntax {
			fmt.Fprintf(writer, "%v\t", indexMap[indexObject.ID.Key()])
		}
		fmt.Fprintf(writer, "|")
		for _, entryObject := range table.EntrySyntax {
			fmt.Fprintf(writer, "\t%v", entryMap[entryObject.ID.Key()])
		}
		fmt.Fprintf(writer, "\n")

		return nil
	}

	if err := client.WalkTable(table, walkRow); err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func main() {
	options.Main(func(args []string) error {
		return options.WithClientID(args, snmptable)
	})
}
