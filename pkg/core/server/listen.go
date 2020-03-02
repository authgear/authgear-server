package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func ListenAndServe(server *http.Server, logger *logrus.Entry, startupMessage string) {
	go func() {
		logger.Info(startupMessage)
		if err := server.ListenAndServe(); err != nil {
			logger.WithError(err).Error("cannot start HTTP server")
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		logger.Info("stopping HTTP server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Fatal("cannot shutdown server")
	}
}
