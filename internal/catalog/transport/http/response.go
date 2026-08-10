package cataloghttp

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type errorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeJSON(w stdhttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(
	w stdhttp.ResponseWriter,
	status int,
	code string,
	message string,
	requestID string,
) {
	writeJSON(w, status, errorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}

func writeCatalogError(
	w stdhttp.ResponseWriter,
	err error,
	requestID string,
) {
	switch {
	case errors.Is(err, catalog.ErrInvalidArgument):
		message := "invalid request"

		var invalidArgumentError *catalog.InvalidArgumentError
		if errors.As(err, &invalidArgumentError) {
			message = invalidArgumentError.Message
		}

		writeError(
			w,
			stdhttp.StatusBadRequest,
			"invalid_argument",
			message,
			requestID,
		)
	case errors.Is(err, catalog.ErrNotFound):
		writeError(
			w,
			stdhttp.StatusNotFound,
			"not_found",
			"resource not found",
			requestID,
		)
	case errors.Is(err, catalog.ErrConflict):
		writeError(
			w,
			stdhttp.StatusConflict,
			"conflict",
			"resource conflict",
			requestID,
		)
	default:
		writeError(
			w,
			stdhttp.StatusInternalServerError,
			"internal",
			"internal server error",
			requestID,
		)
	}
}
