// Package responses re-exports the HTTP response helpers from internal/platform/web
// for transitional legacy code. New modules must import from internal/platform/web directly.
package responses

import "github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"

type HttpResponse = web.HttpResponse

var Ok = web.Ok
var Created = web.Created
var NoContent = web.NoContent
var BadRequest = web.BadRequest
var Unauthorized = web.Unauthorized
var NotFound = web.NotFound
var ServerError = web.ServerError
