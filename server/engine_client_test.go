package server

import (
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"github.com/stretchr/testify/mock"
)

type testEngineClient struct {
	config client.Config

	mock *mock.Mock
}

func (c *testEngineClient) String() string {
	return c.config.String()
}

func (c *testEngineClient) Probe(ids []mibs.ID) ([]bool, error) {
	if c.mock != nil {
		var args = c.mock.MethodCalled("Probe", ids)

		return args.Get(0).([]bool), args.Error(1)
	} else {
		return nil, nil
	}
}

func (c *testEngineClient) WalkObjects(objects []*mibs.Object, f func(*mibs.Object, mibs.IndexValues, mibs.Value, error) error) error {
	return nil // TODO
}

func (c *testEngineClient) WalkTable(table *mibs.Table, f func(mibs.IndexValues, mibs.EntryValues, error) error) error {
	return nil // TODO
}
