package server

import (
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type mibRegistry struct{}

type tableWrapper struct {
	*mibs.Table
}

func (table tableWrapper) makeAPIIndex() api.TableIndex {
	var index = api.TableIndex{
		ID:        table.Table.String(),
		IndexKeys: make([]string, len(table.IndexSyntax)),
		EntryKeys: make([]string, len(table.EntrySyntax)),
	}

	for i, indexObject := range table.IndexSyntax {
		index.IndexKeys[i] = indexObject.String()
	}
	for i, entryObject := range table.EntrySyntax {
		index.EntryKeys[i] = entryObject.String()
	}

	return index
}

func (registry mibRegistry) makeAPIIndex() []api.MIBIndex {
	var index []api.MIBIndex

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		var mibIndex = api.MIBIndex{
			ID:      mib.String(),
			Objects: []api.ObjectIndex{},
			Tables:  []api.TableIndex{},
		}

		mib.Walk(func(id mibs.ID) {
			if object := mib.Object(id); object != nil {
				mibIndex.Objects = append(mibIndex.Objects, api.ObjectIndex{
					ID: object.String(),
				})
			}

			if table := mib.Table(id); table != nil {
				mibIndex.Tables = append(mibIndex.Tables, tableWrapper{table}.makeAPIIndex())
			}
		})

		index = append(index, mibIndex)
	})

	return index
}
