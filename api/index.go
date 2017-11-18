package api

// GET /api/ => { ... }
type Index struct {
	Hosts []HostIndex
	MIBs  []MIBIndex

	IndexObjects
	IndexTables
}
