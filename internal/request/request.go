package request

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// DecodeBody decodes the JSON request body into the target type and runs struct validation.
// If decoding fails, trailing data exists, or validation fails, it writes a Bad Request response and returns false.
func DecodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, response.MsgInvalidBody)
		return req, false
	}
	if dec.More() {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, response.MsgInvalidBody)
		return req, false
	}
	if err := validate.Struct(req); err != nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
			Err:     err,
		})
		return req, false
	}
	return req, true
}

// URLParam returns the URL path parameter with the given key.
// If it is missing, empty, or whitespace, it writes a 400 response and returns false.
func URLParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	v := strings.TrimSpace(chi.URLParam(r, key))
	if v == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		})
		return "", false
	}
	return v, true
}

// QueryParam returns the query parameter with the given key.
// If it is missing, empty, or whitespace, it writes a 400 response and returns false.
func QueryParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		})
		return "", false
	}
	return v, true
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
