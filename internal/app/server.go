package app

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	url2 "net/url"
)

type APIServer struct {
	logger  *logrus.Logger
	router  *http.ServeMux
	storage map[string]string
}

func New() *APIServer {
	return &APIServer{
		logger:  logrus.New(),
		router:  http.NewServeMux(),
		storage: make(map[string]string),
	}
}

func (s *APIServer) Start() error {
	s.configureRouter()

	return http.ListenAndServe(":8080", s.router)
}

func (s *APIServer) configureRouter() {
	s.router.Handle("/", s.myHandler())
}

func shortenURL(url *url2.URL) string {
	return uuid.New().String()
}
