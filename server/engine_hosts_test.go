package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"fmt"
)

func TestEngineAddHosts(t *testing.T) {
	var engine = makeTestEngine()
	var host1 = newHost(HostID("test1"))
	var host2 = newHost(HostID("test2"))

	assert.Truef(t, engine.AddHost(host1), "engine.AddHost %v: to empty hosts", host1)
	assert.Truef(t, engine.AddHost(host2), "engine.AddHost %v: to empty hosts", host2)
	assert.Falsef(t, engine.AddHost(host1), "engine.AddHost %v: to pre-existing hosts", host1)

	assert.Equalf(t, MakeHosts(host1, host2), engine.Hosts(), "engine.Hosts")
}

func TestEngineSetHost(t *testing.T) {
	var engine = makeTestEngine()
	var host1 = newHost(HostID("test"))
	var host2 = newHost(HostID("test"))

	host1.err = fmt.Errorf("test 1")
	host2.err = fmt.Errorf("test 2")

	engine.SetHost(host1)
	engine.SetHost(host2)

	assert.Equalf(t, MakeHosts(host2), engine.Hosts(), "engine.Hosts")
}

func TestEngineDelHost(t *testing.T) {
	var engine = makeTestEngine()
	var host1 = newHost(HostID("test1"))
	var host2 = newHost(HostID("test2"))

	engine.SetHost(host1)
	engine.SetHost(host2)
	engine.DelHost(host1)

	assert.Equalf(t, MakeHosts(host2), engine.Hosts(), "engine.Hosts")
}
