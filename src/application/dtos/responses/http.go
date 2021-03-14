package responses

import (
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

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

func BadRequest(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusBadRequest, formatError(data))
}

func Unauthorized(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusUnauthorized, formatError(customErrors.InvalidTokenMessage))
}

func NotFound(data interface{}) *HttpResponse {
	return newHttpResponse(http.StatusNotFound, formatError(data))
}

func ServerError() *HttpResponse {
	return newHttpResponse(http.StatusInternalServerError, formatError(customErrors.InternalServerError))
}
