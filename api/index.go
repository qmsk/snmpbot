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

type ObjectIndex struct {
	ID string
}

type TableIndex struct {
	ID string

	IndexKeys []string
	EntryKeys []string
}

type HostIndex struct {
	ID   string
	SNMP string
}
