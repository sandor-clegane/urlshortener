package cookie

import "net/http"

var _ CookieService = &cookieServiceImpl{}

type CookieService interface {
	GetUserID(r *http.Request) (string, error)
	createAndSign(w http.ResponseWriter, r *http.Request) error
	checkSign(r *http.Request, name string) error
	Authentication(next http.Handler) http.Handler
}
