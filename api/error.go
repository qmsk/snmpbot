package api

import (
	"encoding/json"
)

type Error struct {
	Error error
}

func (err Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Error.Error())
}
