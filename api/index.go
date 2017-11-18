package api

type Index struct {
	Hosts   []HostIndex
	MIBs    []MIBIndex
	Objects []ObjectIndex
	Tables  []TableIndex
}
