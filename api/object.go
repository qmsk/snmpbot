package api

type ObjectIndex struct {
	ID string
}

type Object struct {
	ObjectIndex
	Value interface{} `json:",omitempty"`
	Error *Error      `json:",omitempty"`
}
