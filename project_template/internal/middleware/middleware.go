// Package middlware is where you would normally place middlwares
// you may expand this overtime into diffrent sub packages
package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Middlewares is used to change middlwares
func Middlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// You should add a logger middlware.
