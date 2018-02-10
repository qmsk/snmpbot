package mibs

import (
	"encoding/json"
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"io"
	"log"
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

func (config MIBConfig) loadMIB() (*MIB, error) {
	if buildMIB, err := config.build(); err != nil {
		return nil, err
	} else {
		return registerMIB(buildMIB), nil
	}
}

func (config MIBConfig) loadObjects(mib *MIB) error {
	for _, objectConfig := range config.Objects {
		if object, err := objectConfig.build(mib); err != nil {
			return fmt.Errorf("Invalid Object %v: %v", objectConfig.Name, err)
		} else {
			mib.registerObject(object)
		}
	}

	return nil
}

func (config MIBConfig) loadTables(mib *MIB) error {
	for _, tableConfig := range config.Tables {
		if table, err := tableConfig.build(mib); err != nil {
			return fmt.Errorf("Invalid Table %v: %v", tableConfig.Name, err)
		} else {
			mib.registerTable(table)

			// setup entry objects IndexSyntax
			for _, entryObject := range table.EntrySyntax {
				entryObject.IndexSyntax = table.IndexSyntax
			}
		}
	}

	return nil
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

	if config.Syntax == "" {

	} else if syntax, err := LookupSyntax(config.Syntax); err != nil {
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

// TODO: support AugmentsEntry => EntryName mapping for augmented table indexes
type TableConfig struct {
	ConfigID
	IndexObjects []string
	EntryObjects []string
}

func (config TableConfig) build(mib *MIB) (Table, error) {
	var table = Table{
		IndexSyntax: make(IndexSyntax, len(config.IndexObjects)),
		EntrySyntax: make(EntrySyntax, 0),
	}

	if id, err := config.resolve(mib); err != nil {
		return table, err
	} else {
		table.ID = id
	}

	for i, indexName := range config.IndexObjects {
		if indexObject, err := ResolveObject(indexName); err != nil {
			return table, fmt.Errorf("Invalid IndexObject %v: %v", indexName, err)
		} else {
			table.IndexSyntax[i] = indexObject
		}
	}

	for _, entryName := range config.EntryObjects {
		if entryObject, err := ResolveObject(entryName); err != nil {
			return table, fmt.Errorf("Unknown EntryObject %v: %v", entryName, err)
		} else if entryObject.NotAccessible {
			continue
		} else {
			table.EntrySyntax = append(table.EntrySyntax, entryObject)
		}
	}

	return table, nil
}

type configWalkFunc func(config MIBConfig, path string) error

func walkJSON(r io.Reader, handler configWalkFunc, path string) error {
	var config MIBConfig

	if err := json.NewDecoder(r).Decode(&config); err != nil {
		return err
	}

	return handler(config, path)
}

func walkFile(file *os.File, handler configWalkFunc) error {
	//log.Printf("Load MIB from file: %v", file.Name())

	switch ext := filepath.Ext(file.Name()); ext {
	case ".json":
		return walkJSON(file, handler, file.Name())
	default:
		return fmt.Errorf("Unknown MIB file extension: %v", ext)
	}
}

func walkPath(path string, handler configWalkFunc) error {
	if file, err := os.Open(path); err != nil {
		return err
	} else if fileInfo, err := file.Stat(); err != nil {
		return err
	} else if fileInfo.IsDir() {
		log.Printf("Load MIBs from directory: %v", path)

		if names, err := file.Readdirnames(0); err != nil {
			return err
		} else {
			for _, name := range names {
				if name[0] == '.' {
					continue
				}

				if err := walkPath(filepath.Join(path, name), handler); err != nil {
					return err
				}
			}
		}
	} else {
		return walkFile(file, handler)
	}

	return nil
}

func walkPathMulti(path string, handlers ...configWalkFunc) error {
	for _, handler := range handlers {
		if err := walkPath(path, handler); err != nil {
			return err
		}
	}

	return nil
}

func Load(path string) error {
	return walkPathMulti(path,
		func(mibConfig MIBConfig, path string) error {
			if mib, err := mibConfig.loadMIB(); err != nil {
				return fmt.Errorf("Failed to load MIB from %v: %v", path, err)
			} else if err := mibConfig.loadObjects(mib); err != nil {
				return fmt.Errorf("Failed to load MIB %v objects from %v: %v", mib, path, err)
			} else {
				log.Printf("Load MIB %v from %v with %d objects", mib, path, len(mib.objects))

				return nil
			}
		},
		func(mibConfig MIBConfig, path string) error {
			if mib, err := ResolveMIB(mibConfig.Name); err != nil {
				return fmt.Errorf("Failed to resolve MIB from %v: %v", path, err)
			} else if err := mibConfig.loadTables(mib); err != nil {
				return fmt.Errorf("Failed to load MIB %v tables from %v: %v", mib, path, err)
			} else {
				log.Printf("Load MIB %v from %v with %d tables", mib, path, len(mib.tables))

				return nil
			}
		},
	)
}
