package api

type IndexHosts struct {
	Hosts []HostIndex
}

type HostIndex struct {
	ID   string
	SNMP string

	MIBs []MIBIndex
}

type Host struct {
	HostIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
