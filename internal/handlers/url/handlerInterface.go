package url

import "net/http"

type URLHandler interface {
	GetAllURL(w http.ResponseWriter, r *http.Request)
	ExpandURL(w http.ResponseWriter, r *http.Request)
	ShortenURL(w http.ResponseWriter, r *http.Request)
	ShortenURLwJSON(w http.ResponseWriter, r *http.Request)
	ShortenSomeURL(w http.ResponseWriter, r *http.Request)

	GetAuthorizationMiddleware() func(next http.Handler) http.Handler
}
