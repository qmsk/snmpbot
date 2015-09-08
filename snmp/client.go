package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "log"
    "os"
    "time"
)

const (
    TIMEOUT     = time.Duration(2) * time.Second
    RETRIES     = 3
)

type Config struct {
    Host        string  `json:host`
    Community   string  `json:community`
}

type Client struct {
    config  Config
    log     *log.Logger

    snmp    *gosnmp.GoSNMP
}

func (self *Client) String() string {
    return fmt.Sprintf("%s", self.snmp.Target)
}

func (self Config) Connect() (*Client, error) {
    client := &Client{
        config: self,
        snmp:   &gosnmp.GoSNMP{
            Target:     self.Host,
            Port:       161,
            Version:    gosnmp.Version2c,
            Community:  self.Community,
            Timeout:    TIMEOUT,
            Retries:    RETRIES,
        },
    }

    if err := client.snmp.Connect(); err != nil {
        return nil, err
    }

    return client, nil
}

func (self *Client) Log() {
    self.log = log.New(os.Stderr, fmt.Sprintf("snmp.Client %v: ", self), 0)
}
