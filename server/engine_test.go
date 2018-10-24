package server

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
)

type testEngine struct {
	hosts engineHosts
}

func (e *testEngine) ClientOptions() client.Options {
	return client.Options{}
}

func (e *testEngine) client(config client.Config) (engineClient, error) {
	var client = testEngineClient{}

	return &client, nil
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

	return c
}

func (e *testEngine) QueryTables(query TableQuery) <-chan TableResult {
	var c = make(chan TableResult)

	defer close(c)

	return c

}

func makeTestEngine() *testEngine {
	var engine = testEngine{
		hosts: makeEngineHosts(),
	}

	return &engine
}

type testEngineClient struct {
}

func (c *testEngineClient) String() string {
	return "<test>"
}

func (c *testEngineClient) Probe(ids []mibs.ID) ([]bool, error) {
	return nil, nil
}

func (c *testEngineClient) WalkObjects(objects []*mibs.Object, f func(*mibs.Object, mibs.IndexValues, mibs.Value, error) error) error {
	return nil
}

func (c *testEngineClient) WalkTable(table *mibs.Table, f func(mibs.IndexValues, mibs.EntryValues, error) error) error {
	return nil
}
