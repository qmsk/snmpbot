package snmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testParseOID = []struct {
	str string
	oid OID
}{
	{"", nil},
	{".", OID{}},
	{".1", OID{1}},
	{".1.3", OID{1, 3}},
}

func TestParseOID(t *testing.T) {
	for _, test := range testParseOID {
		if oid, err := ParseOID(test.str); err != nil {
			t.Errorf("parse OID %v: %v", test.str, err)
		} else {
			assert.Equal(t, test.oid, oid, "ParseOID(%#v)", test.str)
		}
	}
}

var testOIDString = []struct {
	oid OID
	str string
}{
	{nil, ""},
	{OID{}, "."},
	{OID{1}, ".1"},
	{OID{1, 3}, ".1.3"},
}

func TestOIDString(t *testing.T) {
	for _, test := range testOIDString {
		str := test.oid.String()

		assert.Equal(t, test.str, str, "OID(%#v).String()", test.oid)
	}
}

var testOIDIndex = []struct {
	oid   OID
	oid2  OID
	index []int
}{
	{
		OID{1, 3, 6, 1, 6, 3, 1},
		OID{1, 3, 6, 1, 6, 3, 2},
		nil,
	},
	{
		OID{1, 3, 6, 1, 6, 3, 1, 1, 5, 1},
		OID{1, 3, 6, 1, 6, 3, 1, 1, 5, 1},
		[]int{},
	},
	{
		OID{1, 3, 6, 1, 6, 3, 1},
		OID{1, 3, 6, 1, 6, 3, 1, 1, 5, 1},
		[]int{1, 5, 1},
	},
}

func TestOIDIndex(t *testing.T) {
	for _, test := range testOIDIndex {
		index := test.oid.Index(test.oid2)

		assert.Equal(t, test.index, index, "OID(%#v).Index(%#v)", test.oid, test.oid2)
	}
}
