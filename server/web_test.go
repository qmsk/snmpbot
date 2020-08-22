package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/qmsk/go-web/webtest"
	"github.com/qmsk/snmpbot/api"
)

func TestEngineGetIndex(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		hosts: map[HostID]HostConfig{
			HostID("test"): HostConfig{
				SNMP:     "public@localhost",
				Location: "test",
			},
		},
	})

	//
	var apiIndex api.Index
	var testIndex = api.Index{
		Hosts: []api.HostIndex{
			api.HostIndex{
				ID:       "test",
				SNMP:     "public@localhost",
				Online:   true,
				Location: "test",
			},
		},
		MIBs: []api.MIBIndex{
			api.MIBIndex{
				ID: "TEST-MIB",
			},
		},
		IndexObjects: api.IndexObjects{
			Objects: []api.ObjectIndex{
				api.ObjectIndex{
					ID: "TEST-MIB::test",
				},
				api.ObjectIndex{
					ID: "TEST-MIB::testID",
				},
				api.ObjectIndex{
					ID:        "TEST-MIB::testName",
					IndexKeys: []string{"TEST-MIB::testID"},
				},
				api.ObjectIndex{
					ID: "TEST-MIB::testEnum",
				},
			},
		},
		IndexTables: api.IndexTables{
			Tables: []api.TableIndex{
				api.TableIndex{
					ID:         "TEST-MIB::testTable",
					IndexKeys:  []string{"TEST-MIB::testID"},
					ObjectKeys: []string{"TEST-MIB::testName"},
				},
			},
		},
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiIndex,
		},
	})

	assert.Equal(t, testIndex, apiIndex, "response index")
}
