package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

type idTest struct {
	str string
	id  ID
	err string
}

func testResolve(t *testing.T, test idTest) {
	id, err := Resolve(test.str)

	if test.err != "" {
		assert.EqualErrorf(t, err, test.err, "Resolve(%#v)", test.str)
	} else if err != nil {
		t.Errorf("Resolve(%#v): %v", test.str, err)
	} else {
		assert.Equal(t, test.id, id, "Resolve(%#v)", test.str)
	}
}

func testIDString(t *testing.T, test idTest) {
	str := test.id.String()

	assert.Equal(t, test.str, str, "%#v.String()", test.id)
}

func testID(t *testing.T, test idTest) {
	testResolve(t, test)
	testIDString(t, test)
}

func TestResolveInvalidSyntax(t *testing.T) {
	testResolve(t, idTest{
		str: ":x",
		err: "Invalid syntax: :x",
	})
}

func TestResolveInvalidMIB(t *testing.T) {
	testResolve(t, idTest{
		str: "::foo",
		err: "Invalid name without MIB: ::foo",
	})
}

func TestResolveMIBNotFoundError(t *testing.T) {
	testResolve(t, idTest{
		str: "ASDF-MIB",
		err: "MIB not found: ASDF-MIB",
	})
}

func TestResolveNameNotFoundError(t *testing.T) {
	testResolve(t, idTest{
		str: "TEST-MIB::missing",
		err: "TEST-MIB name not found: missing",
	})
}

func TestResolveInvalidName(t *testing.T) {
	testResolve(t, idTest{
		str: "TEST-MIB::.0",
		err: "Invalid syntax: TEST-MIB::.0",
	})
}

func TestResolveInvalidIndex(t *testing.T) {
	testResolve(t, idTest{
		str: "TEST-MIB.0..0",
		err: "Invalid OID part: ",
	})
}

func TestIDMIB(t *testing.T) {
	testID(t, idTest{
		str: "TEST-MIB",
		id:  ID{MIB: TestMIB, OID: snmp.OID{1, 0, 1}},
	})
}

func TestIDMIBIndex(t *testing.T) {
	testID(t, idTest{
		str: "TEST-MIB.2.1",
		id:  ID{MIB: TestMIB, OID: snmp.OID{1, 0, 1, 2, 1}},
	})
}

func TestIDMIBName(t *testing.T) {
	testID(t, idTest{
		str: "TEST-MIB::test",
		id:  ID{MIB: TestMIB, Name: "test", OID: snmp.OID{1, 0, 1, 1, 1}},
	})
}

func TestResolveMIBNameIndex(t *testing.T) {
	testResolve(t, idTest{
		str: "TEST-MIB::test.0",
		id:  ID{MIB: TestMIB, Name: "test", OID: snmp.OID{1, 0, 1, 1, 1, 0}},
	})
}
