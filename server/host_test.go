package server

import (
	"fmt"
	"testing"

	"github.com/qmsk/snmpbot/mibs"
	"github.com/stretchr/testify/assert"
)

func TestLoadHost(t *testing.T) {
	var engine = makeTestEngine(testConfig{})

	var host, err = loadHost(engine, HostID("test"), HostConfig{})

	assert.NoError(t, err, "loadHost")
	assert.Equalf(t, "test", host.String(), "Host.String()")
	assert.Equalf(t, "public@test", host.client.String(), "Host.String()")
}

func TestLoadHostConfig(t *testing.T) {
	var engine = makeTestEngine(testConfig{})

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "public@localhost",
	})

	assert.NoError(t, err, "loadHost")
	assert.Equalf(t, "test", host.String(), "Host.String()")
	assert.Equalf(t, "public@localhost", host.client.String(), "Host.String()")
}

func TestLoadHostConfigLocation(t *testing.T) {
	var engine = makeTestEngine(testConfig{})

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP:     "localhost",
		Location: "testing",
	})

	assert.NoError(t, err, "loadHost")
	assert.Equalf(t, "test", host.String(), "Host.String()")
	assert.Equalf(t, "public@localhost", host.client.String(), "Host.client.String()")
	assert.Equalf(t, "testing", host.Config().Location, "Host.Config.Location")
}

func TestLoadHostConfigClientOptions(t *testing.T) {
	var engine = makeTestEngine(testConfig{})
	var options = engine.ClientOptions()

	options.Community = "private"

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP:          "localhost",
		ClientOptions: &options,
	})

	assert.NoError(t, err, "loadHost")
	assert.Equalf(t, "test", host.String(), "Host.String()")
	assert.Equalf(t, "private@localhost", host.client.String(), "Host.String()")
}

func TestLoadHostConfigError(t *testing.T) {
	var engine = makeTestEngine(testConfig{})

	var _, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "localhost:asdf",
	})

	assert.EqualErrorf(t, err, "parse \"udp+snmp://localhost:asdf\": invalid port \":asdf\" after host", "loadHost ParseConfig")
}

func TestLoadHostClientError(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	engine.mockClient("localhost", fmt.Errorf("Test error"))

	var _, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "localhost",
	})

	assert.EqualErrorf(t, err, "NewClient test: Test error", "loadHost client")
}

func TestLoadHostClientProbeError(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	engine.mockClient("localhost", nil)
	engine.clientMock.On("Probe", []mibs.ID{testMIB.ID}).Return([]bool{false}, fmt.Errorf("Test error"))

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "localhost",
	})

	assert.EqualError(t, err, "Probe test: Test error", "loadHost probe")
	assert.Equalf(t, "test", host.String(), "Host.String()")
	assert.False(t, host.IsUp(), "Host.IsUp")
}

func TestLoadHostClientProbeTrue(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	engine.mockClient("localhost", nil)
	engine.clientMock.On("Probe", []mibs.ID{testMIB.ID}).Return([]bool{true}, nil)

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "localhost",
	})

	assert.NoError(t, err, "loadHost")
	assert.Equal(t, "test", host.String(), "Host.String()")
	assert.True(t, host.IsUp(), "Host.IsUp")
	assert.Equal(t, testMIBs, host.MIBs(), "Host.MIBs()")
}

func TestLoadHostClientProbeFalse(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	engine.mockClient("localhost", nil)
	engine.clientMock.On("Probe", []mibs.ID{testMIB.ID}).Return([]bool{false}, nil)

	var host, err = loadHost(engine, HostID("test"), HostConfig{
		SNMP: "localhost",
	})

	assert.NoError(t, err, "loadHost")
	assert.Equal(t, "test", host.String(), "Host.String()")
	assert.True(t, host.IsUp(), "Host.IsUp")
	assert.Empty(t, host.MIBs(), "Host.MIBs()")
}
