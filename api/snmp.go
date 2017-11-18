package api

type Object struct {
	ObjectIndex
	Value interface{} `json:",omitempty"`
	Error *Error      `json:",omitempty"`
}
