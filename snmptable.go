package main

import (
    "flag"
    "fmt"
    "log"
    "github.com/qmsk/snmpbot/snmp"
)

var (
    snmpLog         bool
    snmpCommunity   string
    snmpTable       string
)

func init() {
    flag.BoolVar(&snmpLog, "snmp-log", false,
        "Log SNMP requests")
    flag.StringVar(&snmpCommunity, "snmp-community", "public",
        "SNMPv2 Community")
    flag.StringVar(&snmpTable, "snmp-table", ".1.3.6.1.2.1.2.2",
        "SNMP table OID")
}

func main() {
    flag.Parse()

    snmpConfigBase := snmp.Config{
        Community:  snmpCommunity,
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
            log.Printf("%s: Config %v: Client %v\n", host, snmpConfig, snmpClient)

            if snmpLog {
                snmpClient.Log()
            }
        }

        // walk table
        table := snmp.If_ifTable

        tableMap, err := snmpClient.GetTableMIB(table)
        if err != nil {
            log.Fatalf("%s: GetTable %v: %s\n", host, table, err)
        }

        // header
        fmt.Printf("%s", table.Index.Name)

        for _, entry := range table.Entry {
            fmt.Printf("\t%s", entry.Name)
        }

        fmt.Printf("\n")

        // data
        for index, entryMap := range tableMap {
            fmt.Printf("%s", index)

            for _, tableEntry := range table.Entry {
                value := entryMap[tableEntry.Name]

                fmt.Printf("\t%v", value)
            }

            fmt.Printf("\n")
        }
    }
}
