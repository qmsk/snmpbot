package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
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
    snmp    *gosnmp.GoSNMP
}

func (self *Client) String() string {
    return fmt.Sprintf("%s@%s", self.snmp.Community, self.snmp.Target)
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
