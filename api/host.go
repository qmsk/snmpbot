package api

type IndexHosts struct {
	Hosts []HostIndex
}

// Shallow host metadata (only MIB IDs)
//
// 	* `GET /api/ => { "Hosts": [ { ... } ] }`
// 	* `GET /api/hosts => { "Hosts": [ { ... } ] }`
// 	* `GET /api/hosts/ => [ { ... } ]`
type HostIndex struct {
	ID       string
	SNMP     string
	Online   bool
	Location string

	MIBs []MIBIndex
}

// Deep host metadata (individual MIB objects/tables)
//
// 	* `GET /api/hosts/:host => { ... }`
type Host struct {
	HostIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
