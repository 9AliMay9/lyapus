package cataloghttp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type environmentCursorPayload struct {
	CreatedAt string `json:"created_at"`
	ID        int64  `json:"id"`
}

func encodeEnvironmentCursor(cursor catalog.EnvironmentCursor) (string, error) {
	payload, err := json.Marshal(environmentCursorPayload{
		CreatedAt: cursor.CreatedAt.UTC().Format(
			"2006-01-02T15:04:05.999999999Z07:00",
		),
		ID: cursor.ID,
	})
	if err != nil {
		return "", fmt.Errorf("encode environment cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeEnvironmentCursor(value string) (catalog.EnvironmentCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return catalog.EnvironmentCursor{}, invalidEnvironmentCursor()
	}

	var decoded environmentCursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return catalog.EnvironmentCursor{}, invalidEnvironmentCursor()
	}

	createdAt, err := time.Parse(time.RFC3339Nano, decoded.CreatedAt)
	if err != nil || decoded.ID < 1 {
		return catalog.EnvironmentCursor{}, invalidEnvironmentCursor()
	}

	return catalog.EnvironmentCursor{
		CreatedAt: createdAt.UTC(),
		ID:        decoded.ID,
	}, nil
}

func invalidEnvironmentCursor() error {
	return &catalog.InvalidArgumentError{
		Message: "cursor is invalid",
	}
}
