package api

type Index struct {
	Hosts []HostIndex
	MIBs  []MIBIndex
}

type MIBIndex struct {
	ID string

	Objects []ObjectIndex
	Tables  []TableIndex
}
