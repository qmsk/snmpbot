package api

import (
	"fmt"
	"encoding/json"
)

type Error struct {
	Error error
}

func (err Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Error.Error())
}

func (err Error) UnmarshalJSON(data []byte) error {
	return fmt.Errorf(string(data))
}
