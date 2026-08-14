package cataloghttp

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

func TestTeamCursorRoundTrip(t *testing.T) {
	original := catalog.TeamCursor{
		CreatedAt: time.Date(
			2026,
			time.August,
			11,
			12,
			34,
			56,
			789,
			time.FixedZone("CST", 8*60*60),
		),
		ID: 42,
	}

	encoded, err := encodeTeamCursor(original)
	if err != nil {
		t.Fatalf("encodeTeamCursor() error = %v", err)
	}
	if encoded == "" {
		t.Fatal("encodeTeamCursor() = empty string")
	}
	if strings.ContainsAny(encoded, "+/=") {
		t.Fatalf("encoded cursor = %q, want raw URL-safe base64", encoded)
	}

	got, err := decodeTeamCursor(encoded)
	if err != nil {
		t.Fatalf("decodeTeamCursor() error = %v", err)
	}
	if got.ID != original.ID {
		t.Fatalf("decoded cursor ID = %d, want %d", got.ID, original.ID)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf(
			"decoded cursor created_at = %s, want %s",
			got.CreatedAt,
			original.CreatedAt,
		)
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf(
			"decoded cursor location = %s, want UTC",
			got.CreatedAt.Location(),
		)
	}
}

func TestDecodeTeamCursorRejectsInvalidValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "not base64",
			value: "not-a-valid-cursor!",
		},
		{
			name:  "not JSON",
			value: base64.RawURLEncoding.EncodeToString([]byte("not JSON")),
		},
		{
			name:  "missing fields",
			value: base64.RawURLEncoding.EncodeToString([]byte(`{}`)),
		},
		{
			name: "non-positive ID",
			value: base64.RawURLEncoding.EncodeToString([]byte(
				`{"created_at":"2026-08-11T04:34:56Z","id":0}`,
			)),
		},
		{
			name: "invalid timestamp",
			value: base64.RawURLEncoding.EncodeToString([]byte(
				`{"created_at":"not-a-time","id":42}`,
			)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeTeamCursor(tt.value)

			if !errors.Is(err, catalog.ErrInvalidArgument) {
				t.Fatalf(
					"decodeTeamCursor() error = %v, want ErrInvalidArgument",
					err,
				)
			}
		})
	}
}
