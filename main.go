package main

import (
    "github.com/jessevdk/go-flags"
    "fmt"
    "net/http"
    "net/url"
    "io"
    "encoding/json"
    "log"
    "os"
    "github.com/qmsk/snmpbot/snmp"
    "sort"
    "strings"
    "sync"
)

const (
    HTTP_LISTEN = ":8085"
)

type Options struct {
    HttpListen      string  `long:"http-listen" description:"HTTP listen address"`
    HttpStatic      string  `long:"http-static" description:"HTTP /static path"`
    SnmpLog         bool    `long:"snmp-log" description:"Log SNMP requests"`
    SnmpTrapListen  string  `long:"snmp-trap-listen" description:"SNMP trap listen address"`

    HostsJson       flags.Filename `short:"H" long:"hosts-json" description:"Path to hosts .json"`

    Args    struct {
        Hosts       []string
    } `positional-args:"yes"`
}

/* Set of active SNMP hosts */
type Host struct {
    Name            string
    snmpClient      *snmp.Client

    Objects         map[string]*snmp.Object
    Tables          map[string]*snmp.Table
}

func (self Host) String() string {
    return self.Name
}

/* Top-level state */
type State struct {
    options     Options
    hosts       map[string]*Host

    httpServer  *http.Server
    trapListen  *snmp.TrapListen
}

func (self *State) init(options Options) {
    self.options = options
    self.hosts = make(map[string]*Host)
}

func (self *State) addHost (name string, snmpConfig snmp.Config) (*Host, error ){
    if snmpClient, err := snmp.Connect(snmpConfig); err != nil {
        return nil, fmt.Errorf("Connect %s: %s\n", snmpConfig, err)

    } else {
        if self.options.SnmpLog {
            snmpClient.Log()
        }

        host := &Host{
            Name:       name,
            snmpClient: snmpClient,

            Objects:    make(map[string]*snmp.Object),
            Tables:     make(map[string]*snmp.Table),
        }

        self.hosts[name] = host

        // discover supported MIBS
        err := host.snmpClient.ProbeMIBs(func(snmpMib *snmp.MIB) {
            host.snmpClient.ProbeMIBObjects(snmpMib, func(snmpObject *snmp.Object){
                host.Objects[snmpObject.String()] = snmpObject
            })

            host.snmpClient.ProbeMIBTables(snmpMib, func(snmpTable *snmp.Table){
                host.Tables[snmpTable.String()] = snmpTable
            })
        })

        if err != nil {
            return nil, fmt.Errorf("Client %s: ProbeTables: %s\n", snmpClient, err)
        }

        return host, nil
    }
}

func (self *State) loadHostsJson (stream io.Reader) error {
    decoder := json.NewDecoder(stream)
    configs := make(map[string]snmp.Config)

    if err := decoder.Decode(&configs); err != nil {
        return err
    }

    // host Clients
    for name, snmpConfig := range configs {
        if host, err := self.addHost(name, snmpConfig); err != nil {
            return fmt.Errorf("%s: %v\n", name, err)
        } else {
            log.Printf("Host %s: %v\n", name, host.snmpClient)
        }
    }

    return nil
}

func (self *State) listenTraps() {
    for trap := range self.trapListen.Listen() {
        log.Printf("listenTraps: %s@%s: %s: %s\n", trap.Agent, trap.SysUpTime, snmp.FormatNotificationType(trap.SnmpTrapOID), trap.Objects)
    }
}

/* Polling collections of items */

// multiple concurrent polls stream items, which are collected into a single map
type Item struct {
    Host        string
    Object      string
    Index       string      // optional
    Value       interface{}
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
func (self *Poll) pollHostTable(host *Host, snmpTable *snmp.Table) {
    // keep collect() waiting
    self.waitGroup.Add(1)
    defer self.waitGroup.Done()

    if tableMap, err := host.snmpClient.GetTable(snmpTable); err != nil {
        log.Printf("%s: GetTable %v: %v\n", host, snmpTable, err)
    } else {
        for index, entry := range tableMap {
            log.Printf("%s: GetTable %v: %v %v\n", host, snmpTable, index, entry)

            self.items <- Item{host.Name, snmpTable.String(), index, entry}
        }
    }
}

func (self *Poll) pollHostObject(host *Host, snmpObject *snmp.Object) {
    // keep collect() waiting
    self.waitGroup.Add(1)
    defer self.waitGroup.Done()

    if value, err := host.snmpClient.GetObject(snmpObject); err != nil {
        log.Printf("%s: GetObject %v: %v\n", host, snmpObject, err)
    } else {
        log.Printf("%s: GetObject %v: %v\n", host, snmpObject, value)

        self.items <- Item{host.Name, snmpObject.String(), "", value}
    }
}

// collect all items from ongoing polls into a map, returning once all polls are complete
func (self *Poll) collect(handler func(item Item)) error {
    // close the items chan once all polls are complete
    go func() {
        self.waitGroup.Wait()
        close(self.items)
    }()

    // collect all items from running polls
    for item := range self.items {
        handler(item)
    }

    // TODO: gather errors
    return nil
}

func (self *State) getSnmpIndex (path []string, params url.Values) (interface{}, error) {
    objects := make(map[string]bool)
    tables := make(map[string]bool)

    var results struct{
        Hosts   []string    `json:"hosts"`
        Objects []string    `json:"objects"`
        Tables  []string    `json:"tables"`
    }

    for hostName, host := range self.hosts {
        results.Hosts = append(results.Hosts, hostName)

        for objectName, _ := range host.Objects {
            objects[objectName] = true
        }
        for tableName, _ := range host.Tables {
            tables[tableName] = true
        }
    }
    for objectName, _ := range objects {
        results.Objects = append(results.Objects, objectName)
    }
    for tableName, _ := range tables {
        results.Tables = append(results.Tables, tableName)
    }

    sort.Strings(results.Hosts)
    sort.Strings(results.Objects)
    sort.Strings(results.Tables)

    return results, nil
}

func (self *State) getSnmpObjects (path []string, params url.Values) (interface{}, error) {
    poll := newPoller()
    results := make(map[string]map[string]interface{})

    for hostName, host := range self.hosts {
        results[hostName] = make(map[string]interface{})

        for objectName, snmpObject := range host.Objects {
            results[hostName][objectName] = nil

            go poll.pollHostObject(host, snmpObject)
        }
    }

    // collect
    return results, poll.collect(func(item Item){
        results[item.Host][item.Object] = item.Value
    })
}

func (self *State) getSnmpTables (path []string, params url.Values) (interface{}, error) {
    poll := newPoller()
    results := make(map[string]map[string]map[string]interface{})

    for hostName, host := range self.hosts {
        results[hostName] = make(map[string]map[string]interface{})

        for tableName, snmpTable := range host.Tables {
            results[hostName][tableName] = make(map[string]interface{})

            go poll.pollHostTable(host, snmpTable)
        }
    }

    // collect
    return results, poll.collect(func(item Item){
        results[item.Host][item.Object][item.Index] = item.Value
    })
}

// HTTP entry point
func (self *State) handleHttp (response http.ResponseWriter, request *http.Request) {
    var responseData interface{}
    var responseErr error

    if request.Method == "GET" {
        if err := request.ParseForm(); err != nil {
            panic(err)
        }
    }

    path := strings.Split(request.URL.Path, "/")

    switch path[0] {
    case "":
        responseData, responseErr = self.getSnmpIndex(path[1:], request.Form)
    case "objects":
        responseData, responseErr = self.getSnmpObjects(path[1:], request.Form)
    case "tables":
        responseData, responseErr = self.getSnmpTables(path[1:], request.Form)
    }

    if responseErr != nil {
        response.Header().Set("Content-Type", "text/plain")
        response.WriteHeader(500)
        fmt.Fprintf(response, "%s", responseErr)
    } else if responseData == nil {
        response.WriteHeader(404)
    } else {
        response.Header().Set("Content-Type", "text/json")

        if err := json.NewEncoder(response).Encode(responseData); err != nil {
            log.Printf("State.handleHttp: json.Encode: %s\n", err)
        }
    }
}

func main () {
    options := Options{
        HttpListen: HTTP_LISTEN,
    }

    if args, err := flags.Parse(&options); err != nil {
        log.Printf("Options: %s\n", err)
        os.Exit(1)
    } else {
        log.Printf("Options: %+v %+v\n", options, args)
    }

    var state State
    state.init(options)

    // snmp hosts from json
    if string(options.HostsJson) == "" {

    } else if file, err := os.Open(string(options.HostsJson)); err != nil {
        log.Printf("Open --hosts-json: %s\n", err)
        os.Exit(1)
    } else if err := state.loadHostsJson(file); err != nil {
        log.Printf("Load --hosts-json=%s: %s\n", options.HostsJson, err)
        os.Exit(2)
    } else {

    }

    // snmp hosts from args
    baseConfig := snmp.Config{}

    for _, hostString := range options.Args.Hosts {
        if config, err := snmp.ParseConfig(hostString, baseConfig); err != nil {
            log.Fatalf("host %v: %s\n", hostString, err)
        } else if host, err := state.addHost(config.Host, config); err != nil {
            log.Fatalf("host %v: %s\n", hostString, err)
        } else {
            log.Printf("Host %s: %v\n", host, host.snmpClient)
        }
    }

    // snmp listen
    if options.SnmpTrapListen == "" {

    } else if trapListen, err := snmp.NewTrapListen(options.SnmpTrapListen); err != nil {
        log.Printf("SNMP TrapListen %s: %s\n", options.SnmpTrapListen, err)
        os.Exit(2)
    } else {
        log.Printf("SNMP TrapListen: %s\n", trapListen)

        state.trapListen = trapListen

        go state.listenTraps()
    }

    // http server
    state.httpServer = &http.Server{
        Addr:   options.HttpListen,
    }

    http.Handle("/snmp/", http.StripPrefix("/snmp/", http.HandlerFunc(state.handleHttp)))

    if options.HttpStatic != "" {
        httpStatic := http.FileServer(http.Dir(options.HttpStatic))

        http.Handle("/static/", http.StripPrefix("/static/", httpStatic))
    }

    // run http
    if err := state.httpServer.ListenAndServe(); err != nil {
        log.Fatalf("Start --http-listen=%s: %s\n", state.options.HttpListen, err)
    }
}
