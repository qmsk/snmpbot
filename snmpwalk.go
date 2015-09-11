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
    snmpRoot        string
)

func init() {
    flag.BoolVar(&snmpLog, "snmp-log", false,
        "Log SNMP requests")
    flag.StringVar(&snmpCommunity, "snmp-community", "public",
        "SNMPv2 Community")
    flag.StringVar(&snmpRoot, "snmp-root", "1.3.6",
        "SNMP root OID")
}

func main() {
    flag.Parse()

    snmpConfigBase := snmp.Config{
        Community:  snmpCommunity,
        Object:     snmpRoot,
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
            if snmpLog {
                snmpClient.Log()
            }
        }

        // resolve
        walkOID := snmp.Resolve(snmpConfig.Object)
        if walkOID == nil {
            log.Fatalf("%s: Resolve %s\n", host, snmpConfig.Object)
        }

        log.Printf("%s: Config %v: Client %v: Walk %v\n", host, snmpConfig, snmpClient, walkOID)

        // walk
        err = snmpClient.WalkTree(walkOID, func(oid snmp.OID, snmpValue interface{}) {
            object := snmp.LookupObject(oid)
            name := oid.String()
            value := snmpValue

            if object != nil {
                name = object.Format(oid)

                if objectValue, err := object.ParseValue(value); err != nil {
                    log.Printf("%s: Invalid %s value: %s\n", host, name, err)
                } else {
                    value = objectValue
                }
            }

            fmt.Printf("%v: %v: %#v\n", host, name, value)
        })
        if err != nil {
            log.Fatalf("%s Walk: %s\n", snmpClient, err)
        }
    }
}
