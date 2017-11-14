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

func TestIDMIB(t *testing.T) {
	testID(t, idTest{
		str: "TEST-MIB",
		id:  ID{MIB: TestMIB, Name: "", OID: snmp.OID{1, 0, 1}},
	})
}
