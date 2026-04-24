// Package web provides shared HTTP response primitives used by all modules.
// This package must not import internal/modules or internal/application to
// avoid import cycles.
package web

import (
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

// HttpResponse is the canonical HTTP response envelope used across all modules.
// Modules import this type from internal/platform/web; the legacy path
// internal/application/dtos/responses remains only for transitional code.
type HttpResponse struct {
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

func newHttpResponse(statusCode int, data interface{}) *HttpResponse {
	return &HttpResponse{StatusCode: statusCode, Data: data}
}

func formatError(message interface{}) map[string]string {
	mapError := make(map[string]string)
	mapError["error"] = fmt.Sprintf("%v", message)
	return mapError
}

func Ok(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusOK, data)
}

func Created(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusCreated, data)
}

func NoContent() *HttpResponse {
	return newHttpResponse(http.StatusNoContent, nil)
}

func BadRequest(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusBadRequest, formatError(data))
}

func Unauthorized(_ interface{}) *HttpResponse {
	return newHttpResponse(http.StatusUnauthorized, formatError(customerrors.InvalidTokenMessage))
}

func NotFound(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusNotFound, formatError(data))
}

func ServerError() *HttpResponse {
	return newHttpResponse(http.StatusInternalServerError, formatError(customerrors.InternalServerError))
}
