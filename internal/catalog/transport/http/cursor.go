package cataloghttp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type teamCursorPayload struct {
	CreatedAt string `json:"created_at"`
	ID        int64  `json:"id"`
}

func encodeTeamCursor(cursor catalog.TeamCursor) (string, error) {
	payload, err := json.Marshal(teamCursorPayload{
		CreatedAt: cursor.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		ID:        cursor.ID,
	})
	if err != nil {
		return "", fmt.Errorf("encode team cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeTeamCursor(value string) (catalog.TeamCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return catalog.TeamCursor{}, invalidTeamCursor()
	}

	var decoded teamCursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return catalog.TeamCursor{}, invalidTeamCursor()
	}

	createdAt, err := time.Parse(time.RFC3339Nano, decoded.CreatedAt)
	if err != nil || decoded.ID < 1 {
		return catalog.TeamCursor{}, invalidTeamCursor()
	}

	return catalog.TeamCursor{
		CreatedAt: createdAt.UTC(),
		ID:        decoded.ID,
	}, nil
}

func invalidTeamCursor() error {
	return &catalog.InvalidArgumentError{
		Message: "cursor is invalid",
	}
}
