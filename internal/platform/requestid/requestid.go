package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

const Header = "X-Request-ID"

type contextKey struct{}

var fallbackSequence atomic.Uint64

func Middleware(next http.Handler) http.Handler {
	return middlewareWithGenerator(next, newID)
}

func middlewareWithGenerator(
	next http.Handler,
	generate func() string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generate()
		w.Header().Set(Header, id)

		ctx := context.WithValue(r.Context(), contextKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}

	return fmt.Sprintf(
		"%x-%x",
		time.Now().UnixNano(),
		fallbackSequence.Add(1),
	)
}
