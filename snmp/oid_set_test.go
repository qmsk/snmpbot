package snmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOIDSetGet(t *testing.T) {
	assert.Equal(t, OID(nil), OIDSet{}.Get(OID{1, 2, 1}))
	assert.Equal(t, OID{}, OIDSet{OID{}}.Get(OID{1, 2, 1}))
	assert.Equal(t, OID{1}, OIDSet{OID{1}}.Get(OID{1, 2, 1}))
	assert.Equal(t, OID(nil), OIDSet{OID{1, 1}}.Get(OID{1, 2, 1}))
}

func TestOIDSetAddEmpty(t *testing.T) {
	var oidSet = OIDSet{}

	oidSet.Add(OID{1})

	assert.Equal(t, oidSet, OIDSet{OID{1}})
}

func TestOIDSetAddOther(t *testing.T) {
	var oidSet = OIDSet{OID{1}}

	oidSet.Add(OID{2})

	assert.Equal(t, oidSet, OIDSet{OID{1}, OID{2}})
}

func TestOIDSetAddSub(t *testing.T) {
	var oidSet = OIDSet{OID{1}}

	oidSet.Add(OID{1, 2, 1})

	assert.Equal(t, oidSet, OIDSet{OID{1}})
}

func TestOIDSetAddSup(t *testing.T) {
	var oidSet = OIDSet{OID{1, 2, 1}}

	oidSet.Add(OID{1})

	assert.Equal(t, oidSet, OIDSet{OID{1}})
}
