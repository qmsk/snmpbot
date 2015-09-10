package main

import (
    "flag"
    "fmt"
    "log"
    "github.com/qmsk/snmpbot/snmp"
)

var (
    snmpCommunity   string
    snmpRoot        string
)

func init() {
    flag.StringVar(&snmpCommunity, "snmp-community", "public",
        "SNMPv2 Community")
    flag.StringVar(&snmpRoot, "snmp-root", "",
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
        }

        // resolve root
        var walkOid snmp.OID

        if snmpRoot != "" {
            walkOid = snmp.ParseOID(snmpRoot)
        } else {
            walkOid = snmp.OID{}
        }

        err = snmpClient.Walk(walkOid, func(oid snmp.OID, value interface{}) {
            fmt.Printf("%v: %v: %#v\n", host, snmp.LookupString(oid), value)
        })
        if err != nil {
            log.Fatalf("%s Walk: %s\n", snmpClient, err)
        }
    }
}
