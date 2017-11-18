package api

type IndexObjects struct {
	Objects []ObjectIndex
}

type ObjectIndex struct {
	ID string
}

type Object struct {
	ObjectIndex
	Value interface{} `json:",omitempty"`
	Error *Error      `json:",omitempty"`
}
