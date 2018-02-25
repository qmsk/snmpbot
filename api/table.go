package api

type IndexTables struct {
	Tables []TableIndex
}

// Table metadata
//
// Different Hosts can have different `Tables` depending on what MIBs were probed.
//
//	* `GET /api/ => { "Tables": [ ... ] }`
// 	* `GET /api/mibs/ => [ { "Tables": [ ... ] } ]`
// 	* `GET /api/mibs/:mib => { "Tables": [ ... ] }`
// 	* `GET /api/hosts/:host/ => { "Tables": [ ... ] }`
// 	* `GET /api/tables => { "Tables": [ ... ] }`
type TableIndex struct {
	ID string

	IndexKeys  []string
	ObjectKeys []string
}

// Table data
//
// The same `Table` can contain `Entries` for multiple different `HostID`s!
//
// 	* `GET /api/tables/ => [ { ... }, ... ]`
// 	* `GET /api/tables/:table => { ... }`
// 	* `GET /api/hosts/:host/tables/ => [ { ... }, ... ]`
// 	* `GET /api/hosts/:host/tables/:table => { ... }`
type Table struct {
	TableIndex

	Entries []TableEntry
	Errors  []TableError `json:",omitempty"`
}

type TableIndexMap map[string]interface{}
type TableObjectsMap map[string]interface{}

type TableEntry struct {
	HostID  string `json:",omitempty"` // XXX: always?
	Index   TableIndexMap
	Objects TableObjectsMap
}

type TableError struct {
	HostID string `json:",omitempty"`
	Error  Error
}

// Optional URL ?query params
//
// Multiple values for the same field are OR, multiple fields are AND.
//
// 	* `GET /api/tables/:table`
// 	* `GET /api/hosts/:host/tables/:table`
type TableQuery struct {
	Hosts []string `schema:"host"`
}

// Optional URL ?query params
//
// Multiple values for the same field are OR, multiple fields are AND.
//
// 	* `GET /api/tables/`
// 	* `GET /api/hosts/:host/tables/`
type TablesQuery struct {
	Hosts  []string `schema:"host"`
	Tables []string `schema:"table"`
}
