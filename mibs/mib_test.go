package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"

	"testing"
)

var (
	TestMIB = registerMIB(makeMIB("TEST-MIB", snmp.OID{1, 0, 1}))

	TestObject = TestMIB.registerObject(Object{
		ID:     ID{MIB: TestMIB, Name: "test", OID: snmp.OID{1, 0, 1, 1, 1}},
		Syntax: DisplayStringSyntax{},
	})
)

func TestResolveMIB(t *testing.T) {
	if mib, err := ResolveMIB("TEST-MIB"); err != nil {
		t.Fatalf("ResolveMIB: %v", err)
	} else {
		assert.Equal(t, TestMIB, mib)
	}
}

func TestResolveMIBError(t *testing.T) {
	if _, err := ResolveMIB("ASDF-MIB"); err == nil {
		t.Fatalf("ResolveMIB: %v", err)
	} else {
		assert.EqualError(t, err, "MIB not found: ASDF-MIB")
	}
}

func TestWalkMIBs(t *testing.T) {
	var found = false

	WalkMIBs(func(mib *MIB) {
		if mib == TestMIB {
			found = true
		}
	})

	assert.True(t, found)
}

func TestWalkObjects(t *testing.T) {
	var found = false

	WalkObjects(func(object *Object) {
		if object == TestObject {
			found = true
		}
	})

	assert.True(t, found)
}

func TestResolveObject(t *testing.T) {
	if object, err := ResolveObject("TEST-MIB::test"); err != nil {
		t.Fatalf("ResolveObject: %v", err)
	} else {
		assert.Equal(t, TestObject, object)
	}
}

func TestResolveObjectErrorResolve(t *testing.T) {
	_, err := ResolveObject("ASDF-MIB::test")

	assert.EqualError(t, err, "MIB not found: ASDF-MIB")
}

func TestResolveObjectErrorMIB(t *testing.T) {
	_, err := ResolveObject(".0.1")

	assert.EqualError(t, err, "No MIB for name: .0.1")
}

func TestResolveObjectErrorObject(t *testing.T) {
	_, err := ResolveObject("TEST-MIB.2.1")

	assert.EqualError(t, err, "Not an object: TEST-MIB.2.1")
}

func TestResolveTableErrorResolve(t *testing.T) {
	_, err := ResolveTable("ASDF-MIB::test")

	assert.EqualError(t, err, "MIB not found: ASDF-MIB")
}

func TestResolveTableErrorMIB(t *testing.T) {
	_, err := ResolveTable(".0.1")

	assert.EqualError(t, err, "No MIB for name: .0.1")
}

func TestResolveTableErrorObject(t *testing.T) {
	_, err := ResolveTable("TEST-MIB.2.1")

	assert.EqualError(t, err, "Not a table: TEST-MIB.2.1")
}
