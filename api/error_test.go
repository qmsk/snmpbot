package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestStruct struct {
	Message Error `json:"error"`
}

func TestErrorJSONMarshal(t *testing.T) {
	ts := TestStruct{Message: Error{fmt.Errorf("Test error")}}
	var testData []byte
	var err error
	testData, err = json.Marshal(ts)

	if err != nil {
		t.Errorf("Failed to marshal Error")
	}

	if string(testData) != "{\"error\":\"Test error\"}" {
		t.Errorf("Unexpected JSON marshalled string %s", string(testData))
	}
}

func TestErrorJsonUnmarshal(t *testing.T) {
	var testData []byte = []byte("{\"error\": \"Test error\"}")
	var target TestStruct
	err := json.Unmarshal(testData, &target)
	if err != nil {
		t.Errorf("Failed to unmarshal error")
	}
	if target.Message.Error.Error() != "Test error" {
		t.Errorf("Error message not same as in JSON")
	}
}
