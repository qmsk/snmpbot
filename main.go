package main

import (
    "github.com/qmsk/snmpbot/snmp"
    "github.com/jessevdk/go-flags"
    "encoding/json"
    "log"
    "os"
)

type options struct {
    Args    struct {
        HostsJson   flags.Filename `short:"H" long:"hosts-json" description:"Path to hosts .json"`
    } `positional-args:"yes" required:"yes"`
}

/* Set of active SNMP hosts */
type Hosts struct {
    Hosts map[string]*snmp.Client
}

func (self *Hosts) loadJsonFile (file *os.File) error {
    decoder := json.NewDecoder(file)
    configs := make(map[string]snmp.Config)

    if err := decoder.Decode(&configs); err != nil {
        return err
    }

    // host Clients
    self.Hosts = make(map[string]*snmp.Client)

    for name, config := range configs {
        if client, err := config.Connect(); err != nil {
            log.Printf("Client %s: Connect %s: %s\n", name, config, err)
        } else {
            log.Printf("Client %s: Connect %s\n", name, config)

            self.Hosts[name] = client
        }
    }

    return nil
}

func main () {
    var options options
    var hosts Hosts

    if args, err := flags.Parse(&options); err != nil {
        //log.Printf("Options: %s\n", err)
        os.Exit(1)
    } else {
        log.Printf("Options: %+v %+v\n", options, args)
    }

    if file, err := os.Open((string)(options.Args.HostsJson)); err != nil {
        log.Printf("Open --hosts-json: %s\n", err)
        os.Exit(1)
    } else if err := hosts.loadJsonFile(file); err != nil {
        log.Printf("Load --hosts-json=%s: %s\n", options.Args.HostsJson, err)
        os.Exit(2)
    } else {
        log.Printf("Hosts: %+v\n", hosts)
    }

    for hostName, hostClient := range hosts.Hosts {
        if interfaces, err := hostClient.Interfaces(); err != nil {
            log.Printf("Host %s: SNMP Interfaces: %s\n", hostName, err)
        } else {
            log.Printf("Host %s: SNMP Interfaces:\n", hostName)

            for _, iface := range interfaces {
                log.Printf("\t%3d: %s\n", iface.IfIndex, iface.IfDescr)
            }
        }
    }
}
