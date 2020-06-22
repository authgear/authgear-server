package httputil

import (
	"context"
	"errors"
	"github.com/skygeario/skygear-server/pkg/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Logger *log.Logger
	Server *http.Server
}

func NewServer(loggerFactory *log.Factory, server *http.Server) *Server {
	return &Server{
		Logger: loggerFactory.New("server"),
		Server: server,
	}
}

func (s *Server) ListenAndServe(startupMessage string) {
	go func() {
		s.Logger.Info(startupMessage)
		err := s.Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.WithError(err).Fatal("failed to start HTTP server")
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		s.Logger.Info("stopping HTTP server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Server.Shutdown(ctx)
	if err != nil {
		s.Logger.WithError(err).Fatal("failed to shutdown server")
	}
}
