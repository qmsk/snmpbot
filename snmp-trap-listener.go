package main

import (
    "flag"
    "fmt"
    "log"
    "github.com/qmsk/snmpbot/snmp"
)

var (
    optionsLog         bool
    optionsListen      string
)

func init() {
    flag.BoolVar(&optionsLog, "log", false,
        "Log SNMP messages")
    flag.StringVar(&optionsListen, "listen", ":162",
        "SNMP trap listen address")
}

func main() {
    flag.Parse()

    // snmp listen
    snmpTrapListen, err := snmp.NewTrapListen(optionsListen)
    if err != nil {
        log.Fatalf("SNMP TrapListen: --listen=%s: %s\n", optionsListen, err)
    }

    if optionsLog {
        snmpTrapListen.Log()
    }

    // listen traps
    log.Printf("SNMP TrapListen: %s\n", snmpTrapListen)

    for trap := range snmpTrapListen.Listen() {
        fmt.Printf("%s @%s %s\n", trap.Agent, trap.SysUpTime, snmp.FormatNotificationType(trap.SnmpTrapOID))

        for _, varBind := range trap.Objects {
            oid := snmp.OID(varBind.Name)
            object := snmp.LookupObject(oid)
            name := oid.String()
            value := varBind.Value

            if object != nil {
                name = object.Format(oid)

                if objectValue, err := object.ParseValue(varBind.Value); err != nil {

                } else {
                    value = objectValue
                }
            }

            fmt.Printf("\t%v %v\n", name, value)
        }
        fmt.Printf("\n")
    }
}
