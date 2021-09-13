package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/stepan2volkov/urlshortener/app/config"
)

type Server struct {
	srv http.Server
}

// NewServer creates http.Server with settings from config.Config
func NewServer(conf config.Config, h http.Handler) *Server {
	s := &Server{}
	s.srv = http.Server{
		Addr:              conf.Addr,
		Handler:           h,
		ReadTimeout:       time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(conf.WriteTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(conf.ReadHeaderTimeout) * time.Second,
	}
	return s
}

func (s *Server) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
	cancel()
}
