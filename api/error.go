package api

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Error error
}

// Custom JSON Marshal function for error type
// Formats errors to strings
func (err Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Error.Error())
}

// Custom JSON Unmarshal function for error type
// error string back to Error
func (err *Error) UnmarshalJSON(data []byte) error {
	var errorMessage string

	if errors := json.Unmarshal(data, &errorMessage); errors != nil {
		return errors
	}

	err.Error = fmt.Errorf(errorMessage)

	return nil
}
