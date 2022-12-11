package db

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

type dbHandlerImpl struct {
	address string
}

func NewDBHandler(adrs string) DBHandler {
	return &dbHandlerImpl{
		address: adrs,
	}
}

// PingConnectionDB Добавьте в сервис хендлер GET /ping, который при запросе проверяет соединение с базой данных.
//При успешной проверке хендлер должен вернуть HTTP-статус 200 OK,
//при неуспешной — 500 Internal Server Error
func (h *dbHandlerImpl) PingConnectionDB(w http.ResponseWriter, _ *http.Request) {
	db, err := sql.Open("postgres", h.address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
