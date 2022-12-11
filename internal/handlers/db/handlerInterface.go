package db

import "net/http"

type DBHandler interface {
	PingConnectionDB(w http.ResponseWriter, r *http.Request)
}
