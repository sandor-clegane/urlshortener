package db

import "net/http"

var _ DBHandler = &dbHandlerImpl{}

type DBHandler interface {
	PingConnectionDB(w http.ResponseWriter, r *http.Request)
}
