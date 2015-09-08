package main

import (
    "github.com/jessevdk/go-flags"
    "net/http"
    "io"
    "encoding/json"
    "log"
    "os"
    "github.com/qmsk/snmpbot/snmp"
)

const (
    HTTP_LISTEN = ":8085"
)

type options struct {
    HttpListen      string `long:"http-listen" description:"HTTP listen address"`

    Args    struct {
        HostsJson   flags.Filename `short:"H" long:"hosts-json" description:"Path to hosts .json"`
    } `positional-args:"yes" required:"yes"`
}

/* Set of active SNMP hosts */
type Host struct {
    snmpClient  *snmp.Client

    Interfaces  snmp.InterfaceTable
    BridgeFdb   snmp.Bridge_FdbTable
}

type State struct {
    hosts       map[string]*Host

    httpServer  *http.Server
}

func (self *State) loadHostsJson (stream io.Reader) error {
    decoder := json.NewDecoder(stream)
    configs := make(map[string]snmp.Config)

    if err := decoder.Decode(&configs); err != nil {
        return err
    }

    // host Clients
    self.hosts = make(map[string]*Host)

    for name, config := range configs {
        if client, err := config.Connect(); err != nil {
            log.Printf("Client %s: Connect %s: %s\n", name, config, err)
        } else {
            log.Printf("Client %s: Connect %s\n", name, config)

            self.hosts[name] = &Host{
                snmpClient: client,

                Interfaces: make(snmp.InterfaceTable),
                BridgeFdb: make(snmp.Bridge_FdbTable),
            }
        }
    }

    return nil
}

// http entry point
type hostEntry struct{
    Interfaces  []*snmp.InterfaceEntry
    BridgeFdb   []*snmp.Bridge_FdbEntry
}

func (self *State) handleHttp (response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "text/json")

    hosts := make(map[string]hostEntry)

    for hostName, host := range self.hosts {
        var hostEntry hostEntry

        // update
        host.snmpClient.GetTable(&host.Interfaces)
        host.snmpClient.GetTable(&host.BridgeFdb)

        for _, entry := range host.Interfaces {
            hostEntry.Interfaces = append(hostEntry.Interfaces, entry)
        }
        for _, entry := range host.BridgeFdb {
            hostEntry.BridgeFdb = append(hostEntry.BridgeFdb, entry)
        }

        hosts[hostName] = hostEntry
    }

    if err := json.NewEncoder(response).Encode(hosts); err != nil {
        log.Printf("State.handleHttp: json.Encode: %s\n", err)
    }
}

func main () {
    var options options = options{
        HttpListen: HTTP_LISTEN,
    }
    var state State

    if args, err := flags.Parse(&options); err != nil {
        log.Printf("Options: %s\n", err)
        os.Exit(1)
    } else {
        log.Printf("Options: %+v %+v\n", options, args)
    }

    // hosts from json
    if file, err := os.Open((string)(options.Args.HostsJson)); err != nil {
        log.Printf("Open --hosts-json: %s\n", err)
        os.Exit(1)
    } else if err := state.loadHostsJson(file); err != nil {
        log.Printf("Load --hosts-json=%s: %s\n", options.Args.HostsJson, err)
        os.Exit(2)
    } else {
        log.Printf("Hosts: %+v\n", state.hosts)
    }

    // http server
    state.httpServer = &http.Server{
        Addr:   options.HttpListen,
    }

    http.HandleFunc("/", state.handleHttp)

    if err := state.httpServer.ListenAndServe(); err != nil {
        log.Printf("Start --http-listen=%s: %s\n", options.HttpListen, err)
        os.Exit(1)
    }
}
