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
	Location string `json:",omitempty"`

	MIBs []MIBIndex
}

// Optional URL ?query params
//
// 	* `GET /api/hosts/:host`
type HostQuery struct {
	SNMP      string `schema:"snmp"`
	Community string `schema:"community"`
}

//  * `POST /api/hosts/`
type HostParams struct {
	HostQuery
	ID       string `schema:"id"`
	Location string `schema:"location"`
}

// Deep host metadata (individual MIB objects/tables)
//
// 	* `GET /api/hosts/:host => { ... }`
type Host struct {
	HostIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
