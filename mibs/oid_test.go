package mibs

import (
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

type oidTest struct {
	name string
	oid  snmp.OID
	err  string
}

func testParseOID(t *testing.T, test oidTest) {
	oid, err := ParseOID(test.name)
	if test.err != "" {
		assert.EqualErrorf(t, err, test.err, "ParseOID(%#v)", test.name)
	} else if err != nil {
		t.Errorf("ParseOID(%#v): %v", test.name, err)
	} else {
		assert.Equal(t, test.oid, oid, "ParseOID(%#v)", test.name)
	}
}

func testFormatOID(t *testing.T, test oidTest) {
	name := FormatOID(test.oid)

	assert.Equal(t, test.name, name, "FormatOID(%#v)", test.oid)
}

func testParseFormatOID(t *testing.T, test oidTest) {
	testParseOID(t, test)
	testFormatOID(t, test)
}

func TestOIDError(t *testing.T) {
	testParseOID(t, oidTest{
		name: "ASDF",
		err:  "MIB not found: ASDF",
	})
}

func TestOIDEmpty(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: "",
		oid:  nil,
	})
}

func TestOIDRaw(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: ".1.3.6.1",
		oid:  snmp.OID{1, 3, 6, 1},
	})
}

func TestOIDMIB(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: "TEST-MIB",
		oid:  snmp.OID{1, 0, 1},
	})
}

func TestOIDMIBIndex(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: "TEST-MIB.0",
		oid:  snmp.OID{1, 0, 1, 0},
	})
}

func TestOID(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: "TEST-MIB::test",
		oid:  snmp.OID{1, 0, 1, 1, 1},
	})
}

func TestOIDIndex(t *testing.T) {
	testParseFormatOID(t, oidTest{
		name: "TEST-MIB::test.0",
		oid:  snmp.OID{1, 0, 1, 1, 1, 0},
	})
}

func TestLookupMIBNotFound(t *testing.T) {
	mib := LookupMIB(snmp.OID{1, 2, 0})

	assert.Nil(t, mib)
}

func TestLookupMIB(t *testing.T) {
	mib := LookupMIB(snmp.OID{1, 0, 1})

	assert.Equal(t, TestMIB, mib)
}

func TestLookupObjectNotFound(t *testing.T) {
	object := LookupObject(snmp.OID{1, 2, 0})

	assert.Nil(t, object)
}

func TestLookupObjectNotObject(t *testing.T) {
	object := LookupObject(snmp.OID{1, 0, 1, 0, 1})

	assert.Nil(t, object)
}

func TestLookupObject(t *testing.T) {
	object := LookupObject(snmp.OID{1, 0, 1, 1, 1})

	assert.Equal(t, TestObject, object)
}

func TestLookupObjectIndex(t *testing.T) {
	object := LookupObject(snmp.OID{1, 0, 1, 1, 1, 0})

	assert.Equal(t, TestObject, object)
}
