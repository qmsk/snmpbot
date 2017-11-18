package api

type IndexObjects struct {
	Objects []ObjectIndex
}

type ObjectIndex struct {
	ID        string
	IndexKeys []string `json:",omitempty"`
}

type Object struct {
	HostID string
	ObjectIndex
	Index ObjectIndexMap `json:",omitempty"`
	Value interface{}    `json:",omitempty"`
	Error *Error         `json:",omitempty"`
}

type ObjectIndexMap map[string]interface{}
