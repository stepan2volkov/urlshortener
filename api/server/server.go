package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/stepan2volkov/urlshortener/app/config"
)

type Server struct {
	srv    http.Server
	logger *zap.Logger
}

// NewServer creates http.Server with settings from config.Config
func NewServer(conf config.Config, h http.Handler, logger *zap.Logger) *Server {
	s := &Server{logger: logger}
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
			if errors.Is(err, http.ErrServerClosed) {
				s.logger.Info("server was stopped")
			} else {
				s.logger.Error("server was stopped abnormally", zap.Error(err))
			}

		}
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("error while stopping server", zap.Error(err))
	}
	cancel()
}
