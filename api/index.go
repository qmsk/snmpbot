package api

type Index struct {
	Hosts []HostIndex
	MIBs  []MIBIndex

	IndexObjects
	IndexTables
}
