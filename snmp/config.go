package snmp

import (
    "fmt"
    "log"
    "net/url"
)

const (
    COMMUNITY   = "public"
)

// Parse a pseudo-URL config string:
//  [community "@"] Host
func ParseConfig(str string, base Config) (config Config, err error) {
    str = "snmp://" + str

    configUrl, err := url.Parse(str)
    if err != nil {
        return config, err
    }

    if configUrl.User != nil {
        config.Community = configUrl.User.Username()
    } else if base.Community != "" {
        config.Community = base.Community
    } else {
        config.Community = COMMUNITY
    }

    log.Printf("ParseConfig %s: url=%#v\n", str, configUrl)

    config.Host = configUrl.Host

    return config, nil
}

type Config struct {
    Host        string  `json:host`
    Community   string  `json:community`
}

func (self Config) String() string {
    return fmt.Sprintf("%s@%s", self.Community, self.Host)
}
