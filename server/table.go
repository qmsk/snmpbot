package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
)

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

func (view hostTablesView) Index(name string) (web.Resource, error) {
	if name == "" {
		return view, nil
	} else if table, err := view.host.resolveTable(name); err != nil {
		return nil, web.Errorf(400, "%v", err)
	} else {
		return hostTableView{view.host, table}, nil
	}
}

func (view hostTablesView) GetREST() (web.Resource, error) {
	var apiTables []api.Table

	view.host.walkTables(func(table *mibs.Table) {
		apiTable := hostTableView{view.host, table}.query()

		apiTables = append(apiTables, apiTable)
	})

	return apiTables, nil
}
