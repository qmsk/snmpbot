package api

type HostIndex struct {
	ID   string
	SNMP string

	MIBs []string
}

type Host struct {
	HostIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
