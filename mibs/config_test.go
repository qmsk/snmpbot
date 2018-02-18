package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	if err := Load("test/TEST2-MIB.json"); err != nil {
		panic(err)
	}
}

func TestConfigResolveMIB(t *testing.T) {
	if resolveMIB, err := ResolveMIB("TEST2-MIB"); err != nil {
		t.Errorf("ResolveMIB TEST2-MIB: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB", resolveMIB.String())
	}
}

func TestConfigResolve(t *testing.T) {
	if id, err := Resolve("TEST2-MIB"); err != nil {
		t.Errorf("Resolve TEST2-MIB: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB", id.String())
	}
}

func TestConfigResolveObjectID(t *testing.T) {
	if id, err := Resolve("TEST2-MIB::test"); err != nil {
		t.Errorf("Resolve TEST2-MIB::test: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::test", id.String())
	}
}

func TestConfigResolveTableID(t *testing.T) {
	if id, err := Resolve("TEST2-MIB::testTable"); err != nil {
		t.Errorf("Resolve TEST2-MIB::testTable: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::testTable", id.String())
	}
}

func TestConfigResolveObject(t *testing.T) {
	if object, err := ResolveObject("TEST2-MIB::test"); err != nil {
		t.Errorf("ResolveObject TEST2-MIB::test: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::test", object.String())
		assert.Equal(t, &DisplayStringSyntax{}, object.Syntax)
	}
}

func TestConfigResolveTable(t *testing.T) {
	mib := LookupMIB(snmp.OID{1, 0, 2})

	if table, err := ResolveTable("TEST2-MIB::testTable"); err != nil {
		t.Errorf("ResolveTable TEST2-MIB::testTable: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::testTable", table.String())
		assert.Equal(t, IndexSyntax{mib.ResolveObject("testID")}, table.IndexSyntax)
		assert.Equal(t, EntrySyntax{mib.ResolveObject("testName")}, table.EntrySyntax)
	}
}

func TestConfigObjectIndexSyntax(t *testing.T) {
	mib := LookupMIB(snmp.OID{1, 0, 2})

	if object, err := ResolveObject("TEST2-MIB::testName"); err != nil {
		t.Errorf("ResolveObject TEST2-MIB::testName: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::testName", object.String())
		assert.Equal(t, &DisplayStringSyntax{}, object.Syntax)
		assert.Equal(t, IndexSyntax{mib.ResolveObject("testID")}, object.IndexSyntax)

		var varBind = snmp.MakeVarBind(object.OID.Extend(10), []byte("foobar"))

		if name, value, err := object.Format(varBind); err != nil {
			t.Errorf("Object<%v>.Format %v: %v", object, varBind, err)
		} else {
			assert.Equal(t, "TEST2-MIB::testName[10]", name)
			assert.Equal(t, "foobar", fmt.Sprintf("%v", value))
		}
	}
}

func TestConfigObjectEnumSyntax(t *testing.T) {
	if object, err := ResolveObject("TEST2-MIB::testEnum"); err != nil {
		t.Errorf("ResolveObject TEST2-MIB::testEnum: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::testEnum", object.String())
		assert.Equal(t, &EnumSyntax{
			{1, "one"},
			{2, "two"},
		}, object.Syntax)

		var varBind = snmp.MakeVarBind(object.OID.Extend(0), int(1))

		if name, value, err := object.Format(varBind); err != nil {
			t.Errorf("Object<%v>.Format %v: %v", object, varBind, err)
		} else {
			assert.Equal(t, "TEST2-MIB::testEnum", name)
			assert.Equal(t, "one", fmt.Sprintf("%v", value))
		}
	}
}

func TestConfigResolveIDExternal(t *testing.T) {
	if object, err := ResolveObject("TEST2-MIB::extObject"); err != nil {
		t.Errorf("ResolveObject TEST2-MIB::extObject: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::extObject", object.String())
	}
}

func TestConfigLookupObject(t *testing.T) {
	if object := LookupObject(snmp.OID{1, 0, 2, 1, 1}); object == nil {
		t.Errorf("LookupObject .1.0.2.1.1: %v", nil)
	} else {
		assert.Equal(t, "TEST2-MIB::test", object.String())

		var varBind = snmp.MakeVarBind(object.OID.Extend(0), []byte("foobar"))

		if name, value, err := object.Format(varBind); err != nil {
			t.Errorf("Object<%v>.Format %v: %v", object, varBind, err)
		} else {
			assert.Equal(t, "TEST2-MIB::test", name)
			assert.Equal(t, "foobar", fmt.Sprintf("%v", value))
		}
	}
}

func TestConfigLookupObjectExt(t *testing.T) {
	if object := LookupObject(snmp.OID{1, 1, 5, 1}); object == nil {
		t.Errorf("LookupObject .1.1.5.1: %v", nil)
	} else {
		assert.Equal(t, "TEST2-MIB::extObject", object.String())
	}
}

func TestConfigObjectUnknownSyntax(t *testing.T) {
	if object, err := ResolveObject("TEST2-MIB::testUnknownSyntax"); err != nil {
		t.Errorf("ResolveObject TEST2-MIB::testUnknownSyntax: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB::testUnknownSyntax", object.String())
		assert.Equal(t, nil, object.Syntax)

		var varBind = snmp.MakeVarBind(object.OID.Extend(0), int(1))

		if name, value, err := object.Format(varBind); err != nil {
			t.Errorf("Object<%v>.Format %v: %v", object, varBind, err)
		} else {
			assert.Equal(t, "TEST2-MIB::testUnknownSyntax", name)
			assert.Equal(t, "1", fmt.Sprintf("%v", value))
		}
	}
}
