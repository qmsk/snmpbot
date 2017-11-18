package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

type tablesRoute struct {
}

func (route tablesRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return tablesView{}, nil
	} else if table, err := mibs.ResolveTable(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return tableView{table}, nil
	}
}

func (route tablesRoute) makeIndex() api.IndexTables {
	return api.IndexTables{
		Tables: tablesView{}.makeAPIIndex(),
	}
}

func (route tablesRoute) GetREST() (web.Resource, error) {
	return route.makeIndex(), nil
}

type tableView struct {
	*mibs.Table
}

func (view tableView) makeAPIIndex() api.TableIndex {
	var index = api.TableIndex{
		ID:        view.Table.String(),
		IndexKeys: make([]string, len(view.Table.IndexSyntax)),
		EntryKeys: make([]string, len(view.Table.EntrySyntax)),
	}

	for i, indexObject := range view.Table.IndexSyntax {
		index.IndexKeys[i] = indexObject.String()
	}
	for i, entryObject := range view.Table.EntrySyntax {
		index.EntryKeys[i] = entryObject.String()
	}

	return index
}

type tablesView struct{}

func (view tablesView) makeAPIIndex() []api.TableIndex {
	var tables []api.TableIndex

	mibs.Walk(func(id mibs.ID) {
		if table := id.MIB.Table(id); table != nil {
			tables = append(tables, tableView{table}.makeAPIIndex())
		}
	})

	return tables
}

type mibTablesView struct {
	mib *mibs.MIB
}

func (view mibTablesView) makeAPIIndex() []api.TableIndex {
	var tables []api.TableIndex

	view.mib.Walk(func(id mibs.ID) {
		if table := view.mib.Table(id); table != nil {
			tables = append(tables, tableView{table}.makeAPIIndex())
		}
	})

	return tables
}

type hostTableView struct {
	host  *Host
	table *mibs.Table
}

func (view hostTableView) makeAPIEntry(indexMap mibs.IndexMap, entryMap mibs.EntryMap) api.TableEntry {
	var entry = api.TableEntry{
		Index:   make(api.TableIndexMap),
		Objects: make(api.TableObjectsMap),
	}

	for _, indexObject := range view.table.IndexSyntax {
		entry.Index[indexObject.String()] = indexMap[indexObject.Key()]
	}
	for _, entryObject := range view.table.EntrySyntax {
		entry.Objects[entryObject.String()] = entryMap[entryObject.Key()]
	}

	return entry
}

func (view hostTableView) query() api.Table {
	var table = api.Table{
		TableIndex: tableView{view.table}.makeAPIIndex(),
	}

	if err := view.host.walkTable(view.table, func(indexMap mibs.IndexMap, entryMap mibs.EntryMap) error {
		table.Entries = append(table.Entries, view.makeAPIEntry(indexMap, entryMap))

		return nil
	}); err != nil {
		table.Error = &api.Error{err}
	}

	return table
}

func (view hostTableView) GetREST() (web.Resource, error) {
	return view.query(), nil
}

type hostTablesView struct {
	host *Host
}

func (view hostTablesView) query() []api.Table {
	var tables []api.Table

	view.host.walkTables(func(table *mibs.Table) {
		tables = append(tables, hostTableView{view.host, table}.query())
	})

	return tables
}

func (view hostTablesView) GetREST() (web.Resource, error) {
	return view.query(), nil
}
