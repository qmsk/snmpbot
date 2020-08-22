package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/qmsk/go-web/webtest"
	"github.com/qmsk/snmpbot/api"
)

func TestGetTablesIndex(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		mibs: testMIBs,
	})

	//
	var apiIndexTables api.IndexTables
	var testIndexTables = api.IndexTables{
		Tables: []api.TableIndex{
			api.TableIndex{
				ID:         "TEST-MIB::testTable",
				IndexKeys:  []string{"TEST-MIB::testID"},
				ObjectKeys: []string{"TEST-MIB::testName"},
			},
		},
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/tables",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiIndexTables,
		},
	})

	assert.Equal(t, testIndexTables, apiIndexTables, "response index")
}
