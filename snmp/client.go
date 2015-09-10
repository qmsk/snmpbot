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

type Client struct {
    log     *log.Logger

    gosnmp  *gosnmp.GoSNMP
}

func (self Client) String() string {
    return fmt.Sprintf("%s", self.gosnmp.Target)
}

func Connect(config Config) (*Client, error) {
    client := &Client{
        gosnmp:   &gosnmp.GoSNMP{
            Target:     config.Host,
            Port:       161,
            Version:    gosnmp.Version2c,
            Community:  config.Community,
            Timeout:    TIMEOUT,
            Retries:    RETRIES,
        },
    }

    if err := client.gosnmp.Connect(); err != nil {
        return nil, err
    }

    return client, nil
}

func (self *Client) Log() {
    self.log = log.New(os.Stderr, fmt.Sprintf("snmp.Client %v: ", self), 0)
}

func (self *Client) Walk(oid OID, handler func (oid OID, value interface{})) error {
    return self.gosnmp.Walk(oid.String(), func(snmpVar gosnmp.SnmpPDU) error {
        oid := ParseOID(snmpVar.Name)
        snmpValue := snmpVar.Value

        handler(oid, snmpValue)

        return nil
    })
}
