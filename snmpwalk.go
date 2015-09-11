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

        // resolve root
        var walkOid snmp.OID

        if snmpRoot != "" {
            walkOid = snmp.ParseOID(snmpRoot)
        } else {
            walkOid = snmp.OID{}
        }

        err = snmpClient.WalkTree(walkOid, func(oid snmp.OID, snmpValue interface{}) {
            name := snmp.LookupString(oid)
            value := snmpValue

            _, object, _ := snmp.Lookup(oid)

            if object != nil {
                if syntaxValue, err := object.ParseValue(value); err != nil {
                    log.Printf("%s: Invalid %s value: %s\n", host, name, err)
                } else {
                    value = syntaxValue
                }
            }

            fmt.Printf("%v: %v: %#v\n", host, name, value)
        })
        if err != nil {
            log.Fatalf("%s Walk: %s\n", snmpClient, err)
        }
    }
}
