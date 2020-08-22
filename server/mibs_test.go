package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/qmsk/go-web/webtest"
	"github.com/qmsk/snmpbot/api"
)

func TestGetMibsIndex(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		mibs: testMIBs,
	})

	//
	var apiIndexes []api.MIBIndex
	var testMibIndexes = []api.MIBIndex{
		api.MIBIndex{
			ID: "TEST-MIB",
		},
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/mibs/",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiIndexes,
		},
	})

	assert.Equal(t, testMibIndexes, apiIndexes, "response index")
}

func TestGetMibIndex(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		mibs: testMIBs,
	})

	//
	var apiIndex api.MIB
	var testMibIndex = api.MIB{
		MIBIndex: api.MIBIndex{
			ID: "TEST-MIB",
		},
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
			Target: "/mibs/TEST-MIB",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiIndex,
		},
	})

	assert.Equal(t, testMibIndex.MIBIndex, apiIndex.MIBIndex, "response index")
	assert.ElementsMatch(t, testMibIndex.Objects, apiIndex.Objects, "response index objects")
	assert.ElementsMatch(t, testMibIndex.Tables, apiIndex.Tables, "response index tables")
}

func TestGetMibNotFound(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		mibs: testMIBs,
	})

	//
	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/mibs/TEST-MIBX",
		},
		Response: webtest.APIResponse{
			StatusCode: 404,
			Text:       "MIB not found: TEST-MIBX" + "\n",
		},
	})
}
