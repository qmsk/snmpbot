package server

import (
	"github.com/qmsk/go-web"
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/mibs"
	"path"
)

func FilterTableObjects(table *mibs.Table, filters ...string) *mibs.Table {
	var filteredTable = mibs.Table{
		ID:          table.ID,
		IndexSyntax: table.IndexSyntax,
	}

	for _, entryObject := range table.EntrySyntax {
		var match = false
		var name = entryObject.String()

		for _, filter := range filters {
			if matched, _ := path.Match(filter, name); matched {
				match = true
			}
		}

		if match {
			filteredTable.EntrySyntax = append(filteredTable.EntrySyntax, entryObject)
		}
	}

	return &filteredTable
}

type TableID string

func AllTables() Tables {
	var tables = make(Tables)

	mibs.WalkTables(func(table *mibs.Table) {
		tables.add(table)
	})

	return tables
}

func MakeTables(args ...*mibs.Table) Tables {
	var tables = make(Tables)

	for _, arg := range args {
		tables.add(arg)
	}

	return tables
}

type Tables map[TableID]*mibs.Table

func (tables Tables) add(table *mibs.Table) {
	tables[TableID(table.Key())] = table
}

func (tables Tables) IDs() []mibs.ID {
	var ids = make([]mibs.ID, len(tables))

	for _, table := range tables {
		ids = append(ids, table.ID)
	}

	return ids
}

func (tables Tables) Filter(filters ...string) Tables {
	var filtered = make(Tables)

	for tableID, table := range tables {
		var match = false
		var name = table.String()

		for _, filter := range filters {
			if matched, _ := path.Match(filter, name); matched {
				match = true
			}
		}

		if match {
			filtered[tableID] = table
		}
	}

	return filtered
}

func (tables Tables) FilterObjects(filters ...string) Tables {
	var filtered = make(Tables)

	for tableID, table := range tables {
		table = FilterTableObjects(table, filters...)

		if len(table.EntrySyntax) == 0 {
			// no table objects matched, elide
			continue
		}

		filtered[tableID] = table
	}

	return filtered
}

type tablesRoute struct {
	engine *Engine
}

func (route tablesRoute) Index(name string) (web.Resource, error) {
	if name == "" {
		return &tablesHandler{
			engine: route.engine,
			hosts:  route.engine.Hosts(),
			tables: route.engine.Tables(),
		}, nil
	} else if table, err := mibs.ResolveTable(name); err != nil {
		return nil, web.Errorf(404, "%v", err)
	} else {
		return &tableHandler{
			engine: route.engine,
			hosts:  route.engine.Hosts(),
			table:  table,
		}, nil
	}
}

func (route tablesRoute) makeIndex() api.IndexTables {
	return api.IndexTables{
		Tables: tablesView{tables: route.engine.Tables()}.makeAPIIndex(),
	}
}

func (route tablesRoute) GetREST() (web.Resource, error) {
	return route.makeIndex(), nil
}

type tableView struct {
	table *mibs.Table
}

func (view tableView) makeAPIIndex() api.TableIndex {
	var index = api.TableIndex{
		ID:         view.table.String(),
		IndexKeys:  make([]string, len(view.table.IndexSyntax)),
		ObjectKeys: make([]string, len(view.table.EntrySyntax)),
	}

	for i, indexObject := range view.table.IndexSyntax {
		index.IndexKeys[i] = indexObject.String()
	}
	for i, entryObject := range view.table.EntrySyntax {
		index.ObjectKeys[i] = entryObject.String()
	}

	return index
}

func (view tableView) entryFromResult(result TableResult) api.TableEntry {
	var entry = api.TableEntry{
		HostID:  string(result.Host.id),
		Index:   make(api.TableIndexMap),
		Objects: make(api.TableObjectsMap),
	}

	for i, indexObject := range view.table.IndexSyntax {
		entry.Index[indexObject.String()] = result.IndexValues[i]
	}
	for i, entryObject := range view.table.EntrySyntax {
		entry.Objects[entryObject.String()] = result.EntryValues[i]
	}

	return entry
}

func (view tableView) errorFromResult(result TableResult) api.TableError {
	return api.TableError{
		HostID: string(result.Host.id),
		Error:  api.Error{result.Error},
	}
}

type tablesView struct {
	tables Tables
}

func (view tablesView) makeAPIIndex() []api.TableIndex {
	var tables = make([]api.TableIndex, len(view.tables))

	for _, table := range view.tables {
		tables = append(tables, tableView{table}.makeAPIIndex())
	}

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

type tableHandler struct {
	engine *Engine
	hosts  Hosts
	table  *mibs.Table
	params api.TableQuery
}

func (handler *tableHandler) query() api.Table {
	var table = api.Table{
		TableIndex: tableView{handler.table}.makeAPIIndex(),
	}

	for result := range handler.engine.QueryTables(TableQuery{
		Hosts:  handler.hosts,
		Tables: MakeTables(handler.table),
	}) {
		if result.Error != nil {
			table.Errors = append(table.Errors, tableView{result.Table}.errorFromResult(result))
		} else {
			table.Entries = append(table.Entries, tableView{result.Table}.entryFromResult(result))
		}
	}

	return table
}

func (handler *tableHandler) QueryREST() interface{} {
	return &handler.params
}

func (handler *tableHandler) GetREST() (web.Resource, error) {
	if handler.params.Hosts != nil {
		handler.hosts = handler.hosts.Filter(handler.params.Hosts...)
	}
	if handler.params.Objects != nil {
		handler.table = FilterTableObjects(handler.table, handler.params.Objects...)
	}

	return handler.query(), nil
}

type tablesHandler struct {
	engine *Engine
	hosts  Hosts
	tables Tables
	params api.TablesQuery
}

func (handler *tablesHandler) query() []*api.Table {
	var tableMap = make(map[TableID]*api.Table, len(handler.tables))
	var tables = make([]*api.Table, 0, len(handler.tables))

	for tableID, t := range handler.tables {
		var table = &api.Table{
			TableIndex: tableView{t}.makeAPIIndex(),
			Entries:    []api.TableEntry{},
		}

		tableMap[tableID] = table
		tables = append(tables, table)
	}

	for result := range handler.engine.QueryTables(TableQuery{
		Hosts:  handler.hosts,
		Tables: handler.tables,
	}) {
		var table = tableMap[TableID(result.Table.Key())]

		if result.Error != nil {
			table.Errors = append(table.Errors, tableView{result.Table}.errorFromResult(result))
		} else {
			table.Entries = append(table.Entries, tableView{result.Table}.entryFromResult(result))
		}
	}

	return tables
}

func (handler *tablesHandler) QueryREST() interface{} {
	return &handler.params
}

func (handler *tablesHandler) GetREST() (web.Resource, error) {
	if handler.params.Hosts != nil {
		handler.hosts = handler.hosts.Filter(handler.params.Hosts...)
	}
	if handler.params.Tables != nil {
		handler.tables = handler.tables.Filter(handler.params.Tables...)
	}
	if handler.params.Objects != nil {
		handler.tables = handler.tables.FilterObjects(handler.params.Objects...)
	}

	return handler.query(), nil
}
