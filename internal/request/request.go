package request

import (
	"context"
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

// Run handles the standard decode-execute-respond lifecycle for a JSON endpoint.
func Run[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	status int,
	action func(ctx context.Context, req Req) (Res, error),
) {
	req, ok := DecodeBody[Req](w, r)
	if !ok {
		return
	}

	res, err := action(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, status, res)
}
