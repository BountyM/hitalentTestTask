package server

import (
	"context"
	"net/http"
	"time"
)

// Server представляет HTTP-сервер с настройками таймаутов.
type Server struct {
	httpServer *http.Server
}

// Run запускает HTTP-сервер на указанном порту с заданным обработчиком.
// Порт должен включать двоеточие, например ":8080".
func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    time.Minute,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown останавливает сервер грациозно с заданным контекстом.
// Если сервер не был запущен, метод ничего не делает.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
