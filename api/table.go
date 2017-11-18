package api

type IndexTables struct {
	Tables []TableIndex
}

type TableIndex struct {
	ID string

	IndexKeys  []string
	ObjectKeys []string
}

type Table struct {
	TableIndex

	Entries []TableEntry
	Error   *Error `json:",omitempty"`
}

type TableIndexMap map[string]interface{}
type TableObjectsMap map[string]interface{}

type TableEntry struct {
	HostID  string `json:",omitempty"` // XXX: always?
	Index   TableIndexMap
	Objects TableObjectsMap
}
