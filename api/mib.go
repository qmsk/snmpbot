package api

// GET /api/ => { "MIBs": [ { ... } ] }
// GET /api/mibs/ => [ { ... } ]
type MIBIndex struct {
	ID string
}

// GET /api/mibs/:mib => { ... }
type MIB struct {
	MIBIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
