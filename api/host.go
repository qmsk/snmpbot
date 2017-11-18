package api

type IndexHosts struct {
	Hosts []HostIndex
}

// GET /api/ => { "Hosts": [ { ... } ] }
// GET /api/hosts => { "Hosts": [ { ... } ] }
// GET /api/hosts/ => [ { ... } ]
type HostIndex struct {
	ID   string
	SNMP string

	MIBs []MIBIndex
}

// GET /api/hosts/:host => { ... }
type Host struct {
	HostIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
