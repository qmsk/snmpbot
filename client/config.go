package client

import (
	"net/url"
)

type Config struct {
	Options        // overrides community from URL user@
	Address string // host or host:port from URL
	Object  string // optional object from URL /path
}

// Parse a pseudo-URL config string:
//  [community "@"] Host
func ParseConfig(options Options, clientURL string) (Config, error) {
	var config = Config{
		Options: options,
	}

	if parseURL, err := url.Parse("udp+snmp://" + clientURL); err != nil {
		return config, err
	} else {
		return config, config.parseURL(parseURL)
	}
}

func (config *Config) parseURL(configURL *url.URL) error {
	if configURL.User != nil {
		config.Community = configURL.User.Username()
	}

	config.Address = configURL.Host

	if configURL.Path != "" {
		config.Object = configURL.Path[1:]
	} else {
		config.Object = ""
	}

	return nil
}

func (config Config) String() string {
	str := ""

	if config.Community != "" {
		str += config.Community + "@"
	}

	str += config.Address

	if config.Object != "" {
		str += "/" + config.Object
	}

	return str
}
