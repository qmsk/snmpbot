package api

// MIB identifier
//
// 	* `GET /api/ => { "MIBs": [ { ... } ] }`
// 	* `GET /api/mibs/ => [ { ... } ]`
type MIBIndex struct {
	ID string
}

// MIB metadata
//
//	 * `GET /api/mibs/:mib => { ... }`
type MIB struct {
	MIBIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
