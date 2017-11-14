package api

type Index struct {
	Hosts []HostIndex
}

type HostIndex struct {
	ID   string
	SNMP string
}
