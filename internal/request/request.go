package request

import (
	"encoding/json"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/response"
)

// DecodeBody decodes the JSON request body into the target type.
// If decoding fails, it writes a Bad Request response and returns false.
func DecodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return req, false
	}
	return req, true
}
