// Package responses re-exports the HTTP response helpers from pkg/web
// for transitional legacy code. New modules must import from pkg/web directly.
package responses

import "github.com/jailtonjunior94/financialcontrol-api/pkg/web"

type HttpResponse = web.HttpResponse

var Ok = web.Ok
var Created = web.Created
var NoContent = web.NoContent
var BadRequest = web.BadRequest
var Unauthorized = web.Unauthorized
var NotFound = web.NotFound
var ServerError = web.ServerError
