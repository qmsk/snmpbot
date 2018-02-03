package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"io"
	"os"
	"path/filepath"
)

type ConfigID struct {
	OID  string
	Name string
}

func (config ConfigID) resolve(mib *MIB) (ID, error) {
	if oid, err := snmp.ParseOID(config.OID); err != nil {
		return ID{}, err
	} else {
		return ID{MIB: mib, OID: oid, Name: config.Name}, nil
	}
}

type MIBConfig struct {
	OID     string
	Name    string
	Objects []ObjectConfig
	Tables  []TableConfig

	oid snmp.OID
}

func (config MIBConfig) build() (MIB, error) {
	var id = ID{Name: config.Name}

	if oid, err := snmp.ParseOID(config.OID); err != nil {
		return MIB{}, fmt.Errorf("Invalid OID for MIB %v: %v", config.Name, err)
	} else {
		id.OID = oid
	}

	return makeMIB(id), nil
}

type ObjectConfig struct {
	ConfigID
	Syntax        string
	SyntaxOptions json.RawMessage // TODO
	NotAccessible bool
}

func (config ObjectConfig) build(mib *MIB) (Object, error) {
	var object = Object{
		NotAccessible: config.NotAccessible,
	}

	if id, err := config.resolve(mib); err != nil {
		return object, err
	} else {
		object.ID = id
	}

	if syntax, err := LookupSyntax(config.Syntax); err != nil {
		return object, err
	} else {
		object.Syntax = syntax
	}

	if config.SyntaxOptions != nil {
		// the dynamically loaded syntax interfaces are pointer-valued
		if err := json.Unmarshal(config.SyntaxOptions, object.Syntax); err != nil {
			return object, fmt.Errorf("Invalid SyntaxOptions: %v", err)
		}
	}

	return object, nil
}

type TableConfig struct {
	ConfigID
	IndexObjects []string
	EntryObjects []string
}

func (config TableConfig) build(mib *MIB) (Table, error) {
	var table = Table{
		IndexSyntax: make(IndexSyntax, len(config.IndexObjects)),
		EntrySyntax: make(EntrySyntax, len(config.EntryObjects)),
	}

	if id, err := config.resolve(mib); err != nil {
		return table, err
	} else {
		table.ID = id
	}

	for i, indexName := range config.IndexObjects {
		if indexObject := mib.ResolveObject(indexName); indexObject == nil {
			return table, fmt.Errorf("Unknown IndexObject: %v", indexName)
		} else {
			table.IndexSyntax[i] = indexObject
		}
	}

	for i, entryName := range config.EntryObjects {
		if entryObject := mib.ResolveObject(entryName); entryObject == nil {
			return table, fmt.Errorf("Unknown EntryObject: %v", entryName)
		} else {
			table.EntrySyntax[i] = entryObject
		}
	}

	return table, nil
}

func loadMIB(config MIBConfig) (*MIB, error) {
	var mib *MIB

	if buildMIB, err := config.build(); err != nil {
		return nil, err
	} else {
		mib = registerMIB(buildMIB)
	}

	for _, objectConfig := range config.Objects {
		if object, err := objectConfig.build(mib); err != nil {
			return mib, fmt.Errorf("Invalid Object %v: %v", objectConfig.Name, err)
		} else {
			mib.registerObject(object)
		}
	}

	for _, tableConfig := range config.Tables {
		if table, err := tableConfig.build(mib); err != nil {
			return mib, fmt.Errorf("Invalid Table %v: %v", tableConfig.Name, err)
		} else {
			mib.registerTable(table)

			// setup object IndexSyntax
			for _, indexObject := range table.IndexSyntax {
				indexObject.IndexSyntax = table.IndexSyntax
			}

			for _, entryObject := range table.EntrySyntax {
				entryObject.IndexSyntax = table.IndexSyntax
			}
		}
	}

	return mib, nil
}

func LoadJSON(r io.Reader) (*MIB, error) {
	var config MIBConfig

	if err := json.NewDecoder(r).Decode(&config); err != nil {
		return nil, err
	}

	return loadMIB(config)
}

func Load(path string) (*MIB, error) {
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		switch ext := filepath.Ext(path); ext {
		case ".json":
			return LoadJSON(file)
		default:
			return nil, fmt.Errorf("Unknown MIB file extension: %v", ext)
		}
	}
}
