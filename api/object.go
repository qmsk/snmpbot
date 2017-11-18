package api

type IndexObjects struct {
	Objects []ObjectIndex
}

// GET /api/ => { "Objects": [ ... ] }
// GET /api/mibs/ => [ { "Objects": [ ... ] } ]
// GET /api/mibs/:mib => { "Objects": [ ... ] }
// GET /api/hosts/:host/ => { "Objects": [ ... ] }
// GET /api/objects => { "Objects": [ ... ] }
type ObjectIndex struct {
	ID        string
	IndexKeys []string `json:",omitempty"`
}

type ObjectInstance struct {
	HostID string
	Index  ObjectIndexMap `json:",omitempty"`
	Value  interface{}    `json:",omitempty"`
	Error  *Error         `json:",omitempty"`
}

// GET /api/objects/ => [ { ... }, ... ]
// GET /api/objects/:object => { ... }
//
// GET /api/hosts/:host/objects/ => [ { ... }, ... ]
// GET /api/hosts/:host/objects/:object => { ... }
type Object struct {
	ObjectIndex
	Instances []ObjectInstance
}

type ObjectIndexMap map[string]interface{}

// GET /api/objects/:object
// GET /api/hosts/:host/objects/:object
type ObjectQuery struct {
	Hosts []string `schema:"host"`
}

// GET /api/objects/
// GET /api/hosts/:host/objects/
type ObjectsQuery struct {
	Hosts   []string `schema:"host"`
	Objects []string `schema:"object"`
}
