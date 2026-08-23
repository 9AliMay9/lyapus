package cataloghttp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type serviceCursorPayload struct {
	CreatedAt string `json:"created_at"`
	ID        int64  `json:"id"`
}

func encodeServiceCursor(cursor catalog.ServiceCursor) (string, error) {
	payload, err := json.Marshal(serviceCursorPayload{
		CreatedAt: cursor.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		ID:        cursor.ID,
	})
	if err != nil {
		return "", fmt.Errorf("encode service cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeServiceCursor(value string) (catalog.ServiceCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return catalog.ServiceCursor{}, invalidServiceCursor()
	}

	var decoded serviceCursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return catalog.ServiceCursor{}, invalidServiceCursor()
	}

	createdAt, err := time.Parse(time.RFC3339Nano, decoded.CreatedAt)
	if err != nil || decoded.ID < 1 {
		return catalog.ServiceCursor{}, invalidServiceCursor()
	}

	return catalog.ServiceCursor{
		CreatedAt: createdAt.UTC(),
		ID:        decoded.ID,
	}, nil
}

func invalidServiceCursor() error {
	return &catalog.InvalidArgumentError{
		Message: "cursor is invalid",
	}
}
