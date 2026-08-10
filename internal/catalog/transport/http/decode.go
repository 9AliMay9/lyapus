package cataloghttp

import (
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
)

const maxJSONRequestBodyBytes int64 = 1 << 20

func decodeJSONBody(
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
	destination any,
) error {
	r.Body = stdhttp.MaxBytesReader(w, r.Body, maxJSONRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON value")
	}

	return nil
}
