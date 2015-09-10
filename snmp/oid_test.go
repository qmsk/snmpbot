package snmp

import (
    "fmt"
    "testing"
)

func testOID(testOid OID, oid OID) error {
    for i := 0; i < len(oid) || i < len(testOid); i++ {
        if i >= len(oid) {
            return fmt.Errorf("%#V[%v]: short: %v", oid, i, testOid[i])
        } else if i >= len(testOid) {
            return fmt.Errorf("%#V[%v]: long: %v", oid, i, oid[i])
        } else if oid[i] != testOid[i] {
            return fmt.Errorf("%#V[%v] != %v", oid, i, testOid[i])
        }
    }

    return nil
}

var testParseOID = []struct{
    str     string
    oid     OID
}{
    {"",        nil},
    {"1",       OID{1}},
    {".1",      OID{1}},
}

func TestParseOID(t *testing.T) {
    for _, test := range testParseOID {
        oid := parseOID(test.str)

        if err := testOID(test.oid, oid); err != nil {
            t.Errorf("fail parseOID(%#v): %s", test.str, err)
        }
    }
}

var testOIDIndex = []struct{
    oid     OID
    oid2    OID
    index   OID
}{
    {OID{1,3,6,1,6,3,1},        OID{1,3,6,1,6,3,1,1,5,1},   OID{1,5,1}},
    {OID{1,3,6,1,6,3,1,1,5,1},  OID{1,3,6,1,6,3,1,1,5,1},   OID{0}},
}

func TestOIDIndex(t *testing.T) {
    for _, test := range testOIDIndex {
        index := test.oid.Index(test.oid2)

        if err := testOID(test.index, index); err != nil {
            t.Errorf("fail %#v.Index(%#v): %s", test.oid, test.oid2, err)
        }

    }
}
