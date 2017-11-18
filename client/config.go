package client

import (
	"net/url"
	"time"
)

type Config struct {
	Logging   Logging
	Community string
	Addr      string // host or host:port
	OID       string
	Timeout   time.Duration
	Retry     int
	UDP       UDPOptions
	MaxVars   uint
}

// Parse a pseudo-URL config string:
//  [community "@"] Host
func (config *Config) Parse(str string) error {
	str = "snmp://" + str

	if parseURL, err := url.Parse(str); err != nil {
		return err
	} else {
		return config.ParseURL(parseURL)
	}
}

func (config *Config) ParseURL(configURL *url.URL) error {
	if configURL.User != nil {
		config.Community = configURL.User.Username()
	}

	//log.Printf("ParseConfig %s: url=%#v\n", str, configUrl)
	config.Addr = configURL.Host

	if configURL.Path != "" {
		config.OID = configURL.Path[1:]
	} else {
		config.OID = ""
	}

	return nil
}

func (config Config) String() string {
	str := ""

	if config.Community != "" {
		str += config.Community + "@"
	}

	str += config.Addr

	if config.OID != "" {
		str += "/" + config.OID
	}

	return str
}
