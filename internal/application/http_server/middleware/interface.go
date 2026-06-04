package middleware

import "net/http"

type Middleware interface {
	Intercept(next http.Handler) http.Handler
}
