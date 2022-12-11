package cookie

import "net/http"

var _ CookieService = &cookieServiceImpl{}

type CookieService interface {
	ExtractValue(cookie *http.Cookie) (string, error)
	CreateAndSign(w http.ResponseWriter, r *http.Request) error
	CheckSign(r *http.Request, name string) error
	Authentication(next http.Handler) http.Handler
}
