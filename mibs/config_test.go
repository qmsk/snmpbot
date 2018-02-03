package mibs

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigLoadMIB(t *testing.T) {
	if mib, err := Load("test/TEST2-MIB.json"); err != nil {
		t.Fatalf("Load test/TEST2-MIB.json: %v", err)
	} else {
		assert.Equal(t, "TEST2-MIB", mib.String())

		if resolveMib, err := ResolveMIB("TEST2-MIB"); err != nil {
			t.Errorf("ResolveMIB TEST2-MIB: %v", err)
		} else {
			assert.Equal(t, mib, resolveMib)
		}

		if id, err := Resolve("TEST2-MIB"); err != nil {
			t.Errorf("Resolve TEST2-MIB: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB", id.String())
		}

		if id, err := Resolve("TEST2-MIB::test"); err != nil {
			t.Errorf("Resolve TEST2-MIB::test: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB::test", id.String())
		}

		if id, err := Resolve("TEST2-MIB::testTable"); err != nil {
			t.Errorf("Resolve TEST2-MIB::testTable: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB::testTable", id.String())
		}

		if object, err := ResolveObject("TEST2-MIB::test"); err != nil {
			t.Errorf("ResolveObject TEST2-MIB::test: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB::test", object.String())
			assert.Equal(t, &DisplayStringSyntax{}, object.Syntax)
		}

		if table, err := ResolveTable("TEST2-MIB::testTable"); err != nil {
			t.Errorf("ResolveTable TEST2-MIB::testTable: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB::testTable", table.String())
			assert.Equal(t, IndexSyntax{mib.ResolveObject("testID")}, table.IndexSyntax)
			assert.Equal(t, EntrySyntax{mib.ResolveObject("testName")}, table.EntrySyntax)
		}

		if object, err := ResolveObject("TEST2-MIB::testName"); err != nil {
			t.Errorf("ResolveObject TEST2-MIB::testName: %v", err)
		} else {
			assert.Equal(t, "TEST2-MIB::testName", object.String())
			assert.Equal(t, &DisplayStringSyntax{}, object.Syntax)
			assert.Equal(t, IndexSyntax{mib.ResolveObject("testID")}, object.IndexSyntax)
		}

		if object := LookupObject(snmp.OID{1, 0, 2, 1, 1}); object == nil {
			t.Errorf("LookupObject .1.0.2.1.1: %v", nil)
		} else {
			assert.Equal(t, "TEST2-MIB::test", object.String())
		}

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
				assert.Equal(t, "TEST2-MIB::testEnum.0", name)
				assert.Equal(t, "one", fmt.Sprintf("%v", value))
			}
		}
	}
}
