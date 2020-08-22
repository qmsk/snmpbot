package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/qmsk/go-web/webtest"
	"github.com/qmsk/snmpbot/api"
)

func TestEngineGetHosts(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{
			HostID("test"): HostConfig{
				SNMP:     "public@localhost",
				Location: "test",
			},
		},
	})

	//
	var apiHostIndexList []api.HostIndex
	var testHostIndex = api.HostIndex{
		ID:       "test",
		SNMP:     "public@localhost",
		Online:   true,
		Location: "test",
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/hosts/",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiHostIndexList,
		},
	})

	assert.Equal(t, []api.HostIndex{testHostIndex}, apiHostIndexList, "response hosts")
}

func TestEngineGetHost(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{
			HostID("test"): HostConfig{
				SNMP:     "public@localhost",
				Location: "test",
			},
		},
	})

	//
	var apiHostIndex api.HostIndex
	var testHostIndex = api.HostIndex{
		ID:       "test",
		SNMP:     "public@localhost",
		Online:   true,
		Location: "test",
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/hosts/test",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiHostIndex,
		},
	})

	assert.Equal(t, testHostIndex, apiHostIndex, "response host")
}

func TestEngineGetHostDynamic(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{},
	})

	//
	var apiHostIndex api.HostIndex
	var testHostIndex = api.HostIndex{
		ID:     "test",
		SNMP:   "public@localhost",
		Online: true,
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/hosts/test/?snmp=public@localhost",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiHostIndex,
		},
	})

	assert.Equal(t, testHostIndex, apiHostIndex, "response host")
}

func TestEngineGetHostDynamicCommunity(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{},
	})

	//
	var apiHost api.Host
	var testHostIndex = api.HostIndex{
		ID:     "test",
		SNMP:   "private@localhost",
		Online: true,
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/hosts/test/?snmp=localhost&community=private",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiHost,
		},
	})

	assert.Equal(t, testHostIndex, apiHost.HostIndex, "response host")
}

func TestEnginePostHost(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{},
	})

	//
	var apiHostIndex api.HostIndex
	var testHostIndex = api.HostIndex{
		ID:       "test",
		SNMP:     "public@localhost",
		Online:   true,
		Location: "test",
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "POST",
			Target: "/hosts/",
			Object: api.HostPOST{
				ID:       "test",
				SNMP:     "public@localhost",
				Location: "test",
			},
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiHostIndex,
		},
	})

	assert.Equal(t, testHostIndex, apiHostIndex, "response host")
	assert.ElementsMatch(t, []HostID{HostID("test")}, engine.Hosts().Keys(), "engine.Hosts")
}
