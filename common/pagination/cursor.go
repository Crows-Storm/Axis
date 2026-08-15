package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const cursorPrefix = "cur_"

// cursorPayload cursor load: Serialized and then base64 encoded
type cursorPayload struct {
	V  string `json:"v"`  // The value of the sorting field (such as the Unix timestamp string for created_time)
	ID string `json:"id"` // A unique ID for records (used to remove duplicates with the same value)
}

func EncodeCursor(value string, id string) string {
	payload := cursorPayload{V: value, ID: id}
	data, _ := json.Marshal(payload)
	return cursorPrefix + base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(cursor string) (value string, id string, err error) {
	if !strings.HasPrefix(cursor, cursorPrefix) {
		return "", "", fmt.Errorf("invalid cursor format")
	}
	raw, err := base64.StdEncoding.DecodeString(cursor[len(cursorPrefix):])
	if err != nil {
		return "", "", fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var payload cursorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", "", fmt.Errorf("invalid cursor payload: %w", err)
	}
	return payload.V, payload.ID, nil
}
