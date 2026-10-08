package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const cursorPrefix = "cur_"

type cursorPayload struct {
	V  string `json:"v"`
	ID string `json:"id"`
}

func EncodeCursor(value string, id string) string {
	data, err := json.Marshal(cursorPayload{V: value, ID: id})
	if err != nil {
		return ""
	}
	return cursorPrefix + base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(cursor string) (value string, id string, err error) {
	if !strings.HasPrefix(cursor, cursorPrefix) {
		return "", "", fmt.Errorf("%w: missing prefix", ErrInvalidCursor)
	}
	raw, err := base64.StdEncoding.DecodeString(cursor[len(cursorPrefix):])
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}
	var payload cursorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}
	if strings.TrimSpace(payload.ID) == "" {
		return "", "", fmt.Errorf("%w: empty id", ErrInvalidCursor)
	}
	return payload.V, payload.ID, nil
}
