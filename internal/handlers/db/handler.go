package db

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type dbHandlerImpl struct {
	DB *sql.DB
}

func NewDBHandler(address string) DBHandler {
	return &dbHandlerImpl{
		DB: connect(address),
	}
}

func connect(dbAddress string) *sql.DB {
	db, err := sql.Open("postgres", dbAddress)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// PingConnectionDB Добавьте в сервис хендлер GET /ping, который при запросе проверяет соединение с базой данных.
//При успешной проверке хендлер должен вернуть HTTP-статус 200 OK,
//при неуспешной — 500 Internal Server Error
func (h *dbHandlerImpl) PingConnectionDB(w http.ResponseWriter, _ *http.Request) {
	err := h.DB.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
