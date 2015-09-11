package snmp

import (
    "log"
    "net"
    "net/url"
)

const (
    COMMUNITY   = "public"
)

// Parse a pseudo-URL config string:
//  [community "@"] Host
func ParseConfig(str string, baseConfig Config) (config Config, err error) {
    str = "snmp://" + str

    configUrl, err := url.Parse(str)
    if err != nil {
        return config, err
    }

    if configUrl.User != nil {
        config.Community = configUrl.User.Username()
    } else if baseConfig.Community != "" {
        config.Community = baseConfig.Community
    } else {
        config.Community = COMMUNITY
    }

    log.Printf("ParseConfig %s: url=%#v\n", str, configUrl)

    if host, port, err := net.SplitHostPort(configUrl.Host); err == nil {
        config.Host = host
        config.Port = port
    } else if baseConfig.Port != "" {
        config.Host = configUrl.Host
        config.Port = baseConfig.Port
    } else {
        config.Host = configUrl.Host
        config.Port = ""
    }

    if configUrl.Path != "" {
        config.Object = configUrl.Path[1:]
    } else if baseConfig.Object != "" {
        config.Object = baseConfig.Object
    } else {
        config.Object = ""
    }

    return config, nil
}

type Config struct {
    Community   string  `json:community`
    Host        string  `json:host`
    Port        string  `json:port`

    Object      string  `json:object`
}

func (self Config) String() string {
    str := ""

    if self.Community != "" {
        str += self.Community + "@"
    }

    str += self.Host

    if self.Port != "" {
        str += ":" + self.Port
    }

    if self.Object != "" {
        str += "/" + self.Object
    }

    return str
}
