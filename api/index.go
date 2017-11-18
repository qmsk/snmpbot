package api

// Metadata
//
// 	* `GET /api/ => { ... }`
type Index struct {
	Hosts []HostIndex
	MIBs  []MIBIndex

	IndexObjects
	IndexTables
}
