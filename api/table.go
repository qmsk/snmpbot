package api

type TableIndex struct {
	ID string

	IndexKeys []string
	EntryKeys []string
}

type Table struct {
	TableIndex

	Entries []TableEntry `json:",omitempty"`
	Error   *Error       `json:",omitempty"`
}

type TableIndexMap map[string]interface{}
type TableObjectsMap map[string]interface{}

type TableEntry struct {
	Index   TableIndexMap
	Objects TableObjectsMap
}
