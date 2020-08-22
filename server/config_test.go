package server

import (
	"os"
	"testing"

	"github.com/qmsk/snmpbot/mibs"
	"github.com/stretchr/testify/assert"
)

var testMIB *mibs.MIB
var testMIBs MIBs

func init() {
	if file, err := os.Open("test/TEST-MIB.json"); err != nil {
		panic(err)
	} else if mib, err := mibs.LoadMIB(file); err != nil {
		panic(err)
	} else {
		testMIB = mib
	}

	testMIBs = MakeMIBs(testMIB)
}

func TestEngineMIBs(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	assert.Equal(t, []string{"TEST-MIB"}, engine.MIBs().Keys(), "Engine.MIBs()")
}

func TestEngineObjects(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	assert.ElementsMatch(t, []string{"TEST-MIB::test", "TEST-MIB::testID", "TEST-MIB::testName", "TEST-MIB::testEnum"}, engine.Objects().Strings(), "Engine.Objects()")
}

func TestEngineTables(t *testing.T) {
	var engine = makeTestEngine(testConfig{clientMock: true})

	assert.ElementsMatch(t, []string{"TEST-MIB::testTable"}, engine.Tables().Strings(), "Engine.Tables()")
}
