package app

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	url2 "net/url"
)

type APIServer struct {
	cfg     *Config
	logger  *logrus.Logger
	router  *http.ServeMux
	storage map[string]string
}

func New(config *Config) *APIServer {
	return &APIServer{
		cfg:     config,
		logger:  logrus.New(),
		router:  http.NewServeMux(),
		storage: make(map[string]string),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	return http.ListenAndServe(s.cfg.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.cfg.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.Handle("/", s.myHandler())
}

func shortenURL(url *url2.URL) string {
	return uuid.New().String()
}
