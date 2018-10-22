package api

type IndexObjects struct {
	Objects []ObjectIndex
}

// Object metadata
//
// Different Hosts can have different `Objects` depending on what MIBs were probed.
//
// 	* `GET /api/ => { "Objects": [ ... ] }`
// 	* `GET /api/mibs/ => [ { "Objects": [ ... ] } ]`
// 	* `GET /api/mibs/:mib => { "Objects": [ ... ] }`
// 	* `GET /api/hosts/:host/ => { "Objects": [ ... ] }`
// 	* `GET /api/objects => { "Objects": [ ... ] }`
type ObjectIndex struct {
	ID        string
	IndexKeys []string `json:",omitempty"`
}

type ObjectInstance struct {
	HostID string
	Index  ObjectIndexMap `json:",omitempty"`
	Value  interface{}    `json:",omitempty"`
}

type ObjectError struct {
	HostID string
	Index  ObjectIndexMap `json:",omitempty"`
	Value  interface{}    `json:",omitempty"`
	Error  Error
}

// Object data
//
// Normal non-tabular objects will only have a single `Instances` entry without any `Index` field.
//
// The same `Object` can contain `Instances` for multiple different `HostID`s!
//
//	* `GET /api/objects/ => [ { ... }, ... ]`
// 	* `GET /api/objects/:object => { ... }`
//
// 	* `GET /api/hosts/:host/objects/ => [ { ... }, ... ]`
// 	* `GET /api/hosts/:host/objects/:object => { ... }`
type Object struct {
	ObjectIndex
	Instances []ObjectInstance
	Errors    []ObjectError `json:",omitempty"`
}

type ObjectIndexMap map[string]interface{}

// Optional URL ?query params
//
// Multiple values for the same field are OR, multiple fields are AND.
//
// 	* `GET /api/objects/:object`
// 	* `GET /api/hosts/:host/objects/:object`
type ObjectQuery struct {
	Hosts []string `schema:"host"`
}

// Optional URL ?query params
//
// Multiple values for the same field are OR, multiple fields are AND.
//
// 	* `GET /api/objects/`
// 	* `GET /api/hosts/:host/objects/`
type ObjectsQuery struct {
	Hosts   []string `schema:"host"`
	Objects []string `schema:"object"`
	Tables  []string `schema:"table"`
}
