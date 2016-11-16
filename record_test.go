package ligno

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateRecordEmpty(t *testing.T) {
	_ = Record{}
}

func TestSerializeRecordToJSON(t *testing.T) {
	recordTime := time.Now().UTC()
	testData := []Record{
		{
			Time:    recordTime,
			Level:   INFO,
			Message: "some message",
			Context: Ctx{"a": "b"},
			Logger:  nil,
		},
		{
			Time:    recordTime,
			Level:   ERROR,
			Context: Ctx{},
			Logger:  GetLogger(""),
		},
	}
	for _, r := range testData {
		marshaled, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var unmarshaled map[string]interface{}
		json.Unmarshal(marshaled, &unmarshaled)
		if _, ok := unmarshaled["time"]; !ok {
			t.Error("time not found in serialized JSON.")
		}
		if _, ok := unmarshaled["level"]; !ok {
			t.Error("level not found in serialzied JSON.")
		}
		if _, ok := unmarshaled["message"]; !ok {
			t.Error("message not found in serialized JSON.")
		}
		if _, ok := unmarshaled["context"]; !ok {
			t.Error("context not found in serialized JSON.")
		}
		if _, ok := unmarshaled["logger"]; ok {
			t.Error("logger found and was not expected.")
		}
	}
}
