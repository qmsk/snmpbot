package main

import (
    "github.com/jessevdk/go-flags"
    "net/http"
    "io"
    "encoding/json"
    "log"
    "os"
    "github.com/qmsk/snmpbot/snmp"
    "sync"
)

const (
    HTTP_LISTEN = ":8085"
)

type options struct {
    HttpListen      string  `long:"http-listen" description:"HTTP listen address"`
    SnmpLog         bool    `long:"snmp-log" description:"Log SNMP requests"`
    SnmpTrapListen  string  `long:"snmp-trap-listen" description:"SNMP trap listen address"`

    Args    struct {
        HostsJson   flags.Filename `short:"H" long:"hosts-json" description:"Path to hosts .json"`
    } `positional-args:"yes" required:"yes"`
}

/* Top-level state */
type State struct {
    hosts       map[string]*Host

    httpServer  *http.Server
    trapListen  *snmp.TrapListen
}

/* Set of active SNMP hosts */
type Table struct {
    sync.Mutex

    Name        string

    Map         interface{} // snmp.table compatible map
}

type Host struct {
    Name            string
    snmpClient      *snmp.Client

    Tables          map[string]*Table
}

func (self *State) loadHostsJson (options options, stream io.Reader) error {
    decoder := json.NewDecoder(stream)
    configs := make(map[string]snmp.Config)

    if err := decoder.Decode(&configs); err != nil {
        return err
    }

    // host Clients
    self.hosts = make(map[string]*Host)

    for name, snmpConfig := range configs {
        if snmpClient, err := snmpConfig.Connect(); err != nil {
            log.Printf("Client %s: Connect %s: %s\n", name, snmpConfig, err)
        } else {
            log.Printf("Client %s: Connect %s\n", name, snmpConfig)

            if options.SnmpLog {
                snmpClient.Log()
            }

            host := &Host{
                Name:       name,
                snmpClient: snmpClient,

                Tables:     make(map[string]*Table),
            }

            host.registerTable("interfaces", make(snmp.InterfaceTable))
            host.registerTable("bridge-fdb", make(snmp.Bridge_FdbTable))

            self.hosts[name] = host
        }
    }

    return nil
}

func (self *Host) registerTable (name string, tableMap interface{}) {
    self.Tables[name] = &Table{Name: name, Map: tableMap}
}

func (self *State) listenTraps() {
    for trap := range self.trapListen.Listen() {
        log.Printf("listenTraps: %s@%s: %s: %s\n", trap.Agent, trap.SysUpTime, snmp.LookupString(trap.SnmpTrapOID), trap.Objects)
    }
}

/* Polling collections of items */

// multiple concurrent polls stream items, which are collected into a single map
type Item struct {
    Host        string
    Table       string
    Index       string
    Entry       interface{} // struct
}

// track state for multiple ongoing polls
type Poll struct {
    items       chan Item
    waitGroup   sync.WaitGroup
}

func newPoller() *Poll {
    return &Poll{
        items:        make(chan Item, 10),
    }
}

// start polling a host-table in the background
func (self *Poll) pollHostTable(host *Host, table *Table) {
    // XXX: this is the routine that updates/accesses the table map;
    //      should be sync'd with only one concurrent goroutine per Table()
    go func() {
        self.waitGroup.Add(1)
        defer self.waitGroup.Done()

        // we must own the table-map while updating/walking it
        table.Lock()
        defer table.Unlock()

        if err := host.snmpClient.GetTable(table.Map); err != nil {
            return
        }

        snmp.WalkTable(table.Map, func(index string, entry interface{}) {
            self.items <- Item{host.Name, table.Name, index, entry}
        })
    }()
}

// collect all items from ongoing polls into a map, returning once all polls are complete
func (self *Poll) collect() interface{} {
    results := make(map[string]map[string]map[string]interface{})

    // close the items chan once all polls are complete
    go func() {
        self.waitGroup.Wait()
        close(self.items)
    }()

    // collect all items from running polls
    for item := range self.items {
        if results[item.Host] == nil {
            results[item.Host] = make(map[string]map[string]interface{})
        }

        if results[item.Host][item.Table] == nil {
            results[item.Host][item.Table] = make(map[string]interface{})
        }

        results[item.Host][item.Table][item.Index] = item.Entry
    }

    return results
}

// HTTP entry point
func (self *State) handleHttp (response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "text/json")

    // poll all available hosts and tables
    poll := newPoller()

    for _, host := range self.hosts {
        for _, table := range host.Tables {
            poll.pollHostTable(host, table)
        }
    }

    // collect and return results
    hosts := poll.collect()

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

    // snmp hosts from json
    if file, err := os.Open((string)(options.Args.HostsJson)); err != nil {
        log.Printf("Open --hosts-json: %s\n", err)
        os.Exit(1)
    } else if err := state.loadHostsJson(options, file); err != nil {
        log.Printf("Load --hosts-json=%s: %s\n", options.Args.HostsJson, err)
        os.Exit(2)
    } else {
        log.Printf("Hosts: %+v\n", state.hosts)
    }

    // snmp listen
    if options.SnmpTrapListen == "" {

    } else if trapListen, err := snmp.NewTrapListen(options.SnmpTrapListen); err != nil {
        log.Printf("Open --snmp-trap listen=%s: %s\n", options.SnmpTrapListen, err)
        os.Exit(2)
    } else {
        log.Printf("SMP TrapListen: %s\n", trapListen)

        state.trapListen = trapListen

        go state.listenTraps()
    }

    // http server
    state.httpServer = &http.Server{
        Addr:   options.HttpListen,
    }

    http.HandleFunc("/", state.handleHttp)

    // XXX: go http
    if err := state.httpServer.ListenAndServe(); err != nil {
        log.Printf("Start --http-listen=%s: %s\n", options.HttpListen, err)
        os.Exit(1)
    }
}
