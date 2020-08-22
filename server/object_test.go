package server

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/qmsk/go-web/webtest"
	"github.com/qmsk/snmpbot/api"
)

func TestGetObjectsIndex(t *testing.T) {
	var engine = makeTestEngine(testConfig{
		mibs: testMIBs,
	})

	//
	var apiIndexObjects api.IndexObjects
	var testIndexObjects = api.IndexObjects{
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
	}

	webtest.TestAPI(t, webtest.APITest{
		Handler: WebAPI(engine),
		Request: webtest.APIRequest{
			Method: "GET",
			Target: "/objects",
		},
		Response: webtest.APIResponse{
			StatusCode: 200,
			Object:     &apiIndexObjects,
		},
	})

	assert.ElementsMatch(t, testIndexObjects.Objects, apiIndexObjects.Objects, "response index")
}
