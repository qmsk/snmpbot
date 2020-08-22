package snmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOIDSetGet(t *testing.T) {
	assert.Equal(t, OID(nil), MakeOIDSet().Get(OID{1, 2, 1}))
	assert.Equal(t, OID{}, MakeOIDSet(OID{}).Get(OID{1, 2, 1}))
	assert.Equal(t, OID{1}, MakeOIDSet(OID{1}).Get(OID{1, 2, 1}))
	assert.Equal(t, OID(nil), MakeOIDSet(OID{1, 1}).Get(OID{1, 2, 1}))
}

func TestOIDSetAddEmpty(t *testing.T) {
	var oidSet = MakeOIDSet()

	oidSet.Add(MustParseOID(".1"))

	assert.Equal(t, "{.1}", oidSet.String())
}

func TestOIDSetAddOther(t *testing.T) {
	var oidSet = MakeOIDSet()

	oidSet.Add(MustParseOID(".1"))
	oidSet.Add(MustParseOID(".2"))

	assert.Equal(t, "{.1 .2}", oidSet.String())
}

func TestOIDSetAddSub(t *testing.T) {
	var oidSet = MakeOIDSet()

	oidSet.Add(MustParseOID(".1"))
	oidSet.Add(MustParseOID(".1.2.1"))

	assert.Equal(t, "{.1}", oidSet.String())
}

func TestOIDSetAddSup(t *testing.T) {
	var oidSet = MakeOIDSet()

	oidSet.Add(MustParseOID(".1.2.1"))
	oidSet.Add(MustParseOID(".1"))

	assert.Equal(t, "{.1}", oidSet.String())
}
