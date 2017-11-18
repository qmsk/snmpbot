package api

type IndexObjects struct {
	Objects []ObjectIndex
}

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

type Object struct {
	ObjectIndex
	Instances []ObjectInstance
}

type ObjectIndexMap map[string]interface{}
