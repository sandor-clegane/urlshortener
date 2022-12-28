package url

import "net/http"

var _ URLHandler = &URLhandlerImpl{}

type URLHandler interface {
	GetAllURL(w http.ResponseWriter, r *http.Request)
	ExpandURL(w http.ResponseWriter, r *http.Request)
	ShortenURL(w http.ResponseWriter, r *http.Request)
	ShortenURLwJSON(w http.ResponseWriter, r *http.Request)
	ShortenSomeURL(w http.ResponseWriter, r *http.Request)
	DeleteSomeURL(w http.ResponseWriter, r *http.Request)
	PingConnectionDB(w http.ResponseWriter, r *http.Request)

	GetAuthorizationMiddleware() func(next http.Handler) http.Handler
}
