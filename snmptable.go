package main

import (
    "flag"
    "fmt"
    "log"
    "github.com/qmsk/snmpbot/snmp"
)

var (
    configLog         bool
    configCommunity   string
    configTable       string
)

func init() {
    flag.BoolVar(&configLog, "log", false,
        "Log SNMP requests")
    flag.StringVar(&configCommunity, "community", "public",
        "SNMPv2 Community")
    flag.StringVar(&configTable, "table", ".1.3.6.1.2.1.2.2",
        "SNMP table OID")
}

func main() {
    flag.Parse()

    snmpConfigBase := snmp.Config{
        Community:  configCommunity,
        Object:     configTable,
    }

    for _, host := range flag.Args() {
        var snmpConfig snmp.Config
        var snmpClient *snmp.Client
        var err error

        if snmpConfig, err = snmp.ParseConfig(host, snmpConfigBase); err != nil {
            log.Fatalf("Client config %s: %s\n", host, err)
        } else if snmpClient, err = snmp.Connect(snmpConfig); err != nil {
            log.Fatalf("Connect %s: %s\n", snmpConfig, err)
        } else {
            if configLog {
                snmpClient.Log()
            }
        }

        // resolve table
        snmpTable := snmp.ResolveTable(snmpConfig.Object)
        if snmpTable == nil {
            log.Fatalf("%s: ResolveTable %v\n", host, snmpConfig.Object)
        }

        log.Printf("%s: Config %v: Client %v: Table %v\n", host, snmpConfig, snmpClient, snmpTable)

        // walk table
        tableMap, err := snmpClient.GetTableMIB(snmpTable)
        if err != nil {
            log.Fatalf("%s: GetTable %v: %s\n", host, snmpTable, err)
        }

        // print header
        fmt.Printf("%s", snmpTable.Index.Name)

        for _, entry := range snmpTable.Entry {
            fmt.Printf("\t%s", entry.Name)
        }

        fmt.Printf("\n")

        // print rows
        for index, entryMap := range tableMap {
            fmt.Printf("%s", index)

            for _, tableEntry := range snmpTable.Entry {
                value := entryMap[tableEntry.Name]

                fmt.Printf("\t%v", value)
            }

            fmt.Printf("\n")
        }
    }
}
