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
	if oid, err := snmp.ParseOID(config.OID); err != nil {
		return MIB{}, fmt.Errorf("Invalid OID for MIB %v: %v", config.Name, err)
	} else {
		return makeMIB(config.Name, oid), nil
	}

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

type loadContext struct {
	entryMap    map[string]*Table // EntryName => *Table
	augmentsMap map[string]string // EntryName => AugmentsEntry
}

func (config MIBConfig) loadTables(mib *MIB, loadContext loadContext) error {
	for _, tableConfig := range config.Tables {
		if table, err := tableConfig.build(mib); err != nil {
			return fmt.Errorf("Invalid Table %v: %v", tableConfig.Name, err)
		} else {
			loadContext.entryMap[mib.Name+"::"+tableConfig.EntryName] = mib.registerTable(table)
			loadContext.augmentsMap[mib.Name+"::"+tableConfig.EntryName] = tableConfig.AugmentsEntry
		}
	}

	return nil
}

func (config MIBConfig) loadTablesIndex(mib *MIB, loadContext loadContext) error {
	for _, tableConfig := range config.Tables {
		table := mib.ResolveTable(tableConfig.Name)

		// resolve IndexSyntax from augmented entry table
		if tableConfig.AugmentsEntry != "" {
			var augmentsEntry = tableConfig.AugmentsEntry

			// chase any chained augments
			for {
				if nextEntry := loadContext.augmentsMap[augmentsEntry]; nextEntry == "" {
					break
				} else {
					augmentsEntry = nextEntry
				}
			}

			if augmentsTable, ok := loadContext.entryMap[augmentsEntry]; !ok {
				return fmt.Errorf("Invalid table %v::%v AugmentsEntry=%v: not found", mib.Name, tableConfig.Name, tableConfig.AugmentsEntry)
			} else if augmentsTable.IndexSyntax == nil {
				return fmt.Errorf("Invalid table %v::%v AugmentsEntry=%v: no index syntax", mib.Name, tableConfig.Name, tableConfig.AugmentsEntry)
			} else {
				table.IndexSyntax = augmentsTable.IndexSyntax
			}
		}

		// setup entry objects IndexSyntax
		for _, entryObject := range table.EntrySyntax {
			entryObject.IndexSyntax = table.IndexSyntax
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

type TableConfig struct {
	ConfigID
	IndexObjects  []string
	EntryObjects  []string
	EntryName     string
	AugmentsEntry string // map IndexObjects from table with EntryName
}

func (config TableConfig) build(mib *MIB) (Table, error) {
	var table = Table{
		EntrySyntax: make(EntrySyntax, 0),
	}

	if id, err := config.resolve(mib); err != nil {
		return table, err
	} else {
		table.ID = id
	}

	if config.AugmentsEntry == "" {
		table.IndexSyntax = make(IndexSyntax, len(config.IndexObjects))

		for i, indexName := range config.IndexObjects {
			if indexObject, err := ResolveObject(indexName); err != nil {
				return table, fmt.Errorf("Invalid IndexObject %v: %v", indexName, err)
			} else {
				table.IndexSyntax[i] = indexObject
			}
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
		log.Infof("Load MIBs from directory: %v", path)

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

// Load and register multiple MIBs recursively from the given filesystem path.
//
// References to other MIBs are resolved via multiple passes, MIB ordering does not matter.
func Load(path string) error {
	var loadContext = loadContext{
		entryMap:    make(map[string]*Table),
		augmentsMap: make(map[string]string),
	}

	return walkPathMulti(path,
		func(mibConfig MIBConfig, path string) error {
			if mib, err := mibConfig.loadMIB(); err != nil {
				return fmt.Errorf("Failed to load MIB from %v: %v", path, err)
			} else if err := mibConfig.loadObjects(mib); err != nil {
				return fmt.Errorf("Failed to load MIB %v objects from %v: %v", mib, path, err)
			} else {
				log.Infof("Load MIB %v from %v with %d objects", mib, path, len(mib.objects))

				return nil
			}
		},
		func(mibConfig MIBConfig, path string) error {
			if mib, err := ResolveMIB(mibConfig.Name); err != nil {
				return fmt.Errorf("Failed to resolve MIB from %v: %v", path, err)
			} else if err := mibConfig.loadTables(mib, loadContext); err != nil {
				return fmt.Errorf("Failed to load MIB %v tables from %v: %v", mib, path, err)
			} else {
				log.Infof("Load MIB %v from %v with %d tables", mib, path, len(mib.tables))

				return nil
			}
		},
		func(mibConfig MIBConfig, path string) error {
			if mib, err := ResolveMIB(mibConfig.Name); err != nil {
				return fmt.Errorf("Failed to resolve MIB from %v: %v", path, err)
			} else if err := mibConfig.loadTablesIndex(mib, loadContext); err != nil {
				return fmt.Errorf("Failed to load MIB %v tables from %v: %v", mib, path, err)
			} else {
				return nil
			}
		},
	)
}

// Load a single MIB, and return it.
//
// Any other MIBs referred to by this MIB must already have been loaded.
func LoadMIB(r io.Reader) (*MIB, error) {
	var mibConfig MIBConfig
	var loadContext = loadContext{
		entryMap:    make(map[string]*Table),
		augmentsMap: make(map[string]string),
	}

	if err := json.NewDecoder(r).Decode(&mibConfig); err != nil {
		return nil, err
	}

	if mib, err := mibConfig.loadMIB(); err != nil {
		return mib, fmt.Errorf("Failed to load MIB: %v", err)
	} else if err := mibConfig.loadObjects(mib); err != nil {
		return mib, fmt.Errorf("Failed to load MIB %v objects: %v", mib, err)
	} else if err := mibConfig.loadTables(mib, loadContext); err != nil {
		return mib, fmt.Errorf("Failed to load MIB %v tables: %v", mib, err)
	} else if err := mibConfig.loadTablesIndex(mib, loadContext); err != nil {
		return mib, fmt.Errorf("Failed to load MIB %v tables: %v", mib, err)
	} else {
		return mib, nil
	}
}
