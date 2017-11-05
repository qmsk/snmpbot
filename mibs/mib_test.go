package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
)

var (
	TestMIB = RegisterMIB("TEST-MIB", snmp.OID{1, 0, 1})

	TestObject = TestMIB.RegisterObject(TestMIB.Register("test", 1, 1), Object{
		Syntax: DisplayStringSyntax,
	})
)
