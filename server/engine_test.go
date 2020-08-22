package server

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/stretchr/testify/mock"
)

type testConfig struct {
	hosts map[HostID]HostConfig

	clientMock bool
}

type testEngine struct {
	hosts engineHosts

	mock.Mock
}

func makeTestEngine(config testConfig) *testEngine {
	var engine = testEngine{
		hosts: makeEngineHosts(),
	}

	if !config.clientMock {
		engine.On("client", mock.AnythingOfType("client.Config")).Return(nil)
	}

	for id, config := range config.hosts {
		if host, err := loadHost(&engine, id, config); err != nil {
			panic(err)
		} else if !engine.hosts.Add(host) {
			panic("host already added")
		}
	}

	return &engine
}

func (e *testEngine) ClientOptions() client.Options {
	return client.Options{
		Community: "public",
	}
}

func (e *testEngine) mockClient(snmp string, clientErr error) {
	if clientOptions, err := client.ParseConfig(e.ClientOptions(), snmp); err != nil {
		panic(err)
	} else {
		e.On("client", clientOptions).Return(clientErr)
	}
}

func (e *testEngine) client(config client.Config) (engineClient, error) {
	var args = e.Called(config)
	var client = testEngineClient{
		config: config,
	}

	return &client, args.Error(0)
}

func (e *testEngine) MIBs() MIBs {
	return AllMIBs()
}

func (e *testEngine) Objects() Objects {
	return AllObjects()
}

func (e *testEngine) Tables() Tables {
	return AllTables()
}

func (e *testEngine) Hosts() Hosts {
	return e.hosts.Copy()
}

func (e *testEngine) AddHost(host *Host) bool {
	return e.hosts.Add(host)
}

func (e *testEngine) SetHost(host *Host) {
	e.hosts.Set(host)
}

func (e *testEngine) DelHost(host *Host) bool {
	return e.hosts.Del(host)
}

func (e *testEngine) QueryObjects(query ObjectQuery) <-chan ObjectResult {
	var c = make(chan ObjectResult)

	defer close(c)

	// TODO
	return c
}

func (e *testEngine) QueryTables(query TableQuery) <-chan TableResult {
	var c = make(chan TableResult)

	defer close(c)

	// TODO
	return c
}
