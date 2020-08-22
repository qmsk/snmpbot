package server

import (
	"os"

	"github.com/qmsk/snmpbot/mibs"
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
