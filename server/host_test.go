package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
