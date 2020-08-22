package server

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
)

type testEngineClient struct {
	config client.Config
}

func (c *testEngineClient) String() string {
	return c.config.String()
}

func (c *testEngineClient) Probe(ids []mibs.ID) ([]bool, error) {
	return nil, nil // TODO
}

func (c *testEngineClient) WalkObjects(objects []*mibs.Object, f func(*mibs.Object, mibs.IndexValues, mibs.Value, error) error) error {
	return nil // TODO
}

func (c *testEngineClient) WalkTable(table *mibs.Table, f func(mibs.IndexValues, mibs.EntryValues, error) error) error {
	return nil // TODO
}
