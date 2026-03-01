package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Router interface {
	RegisterRoutes(mux *http.ServeMux)
}

// SetupRoutes настраивает все маршруты приложения
func SetupRoutes(handler *DepartmentHandler, log *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Создаём middleware с нашим логгером
	logging := LoggingMiddleware(log)

	// POST /departments/ - Создать подразделение
	mux.Handle("POST /departments/", logging(http.HandlerFunc(handler.CreateDepartment)))

	// POST /departments/{id}/employees/
	mux.Handle("POST /departments/{id}/employees/", logging(http.HandlerFunc(handler.CreateEmployee)))

	// GET /departments/{id}
	mux.Handle("GET /departments/{id}", logging(http.HandlerFunc(handler.GetDepartment)))

	// PATCH /departments/{id}
	mux.Handle("PATCH /departments/{id}", logging(http.HandlerFunc(handler.UpdateDepartment)))

	// DELETE /departments/{id}
	mux.Handle("DELETE /departments/{id}", logging(http.HandlerFunc(handler.DeleteDepartment)))

	return mux
}

// responseLogger оборачивает ResponseWriter для захвата статуса
type responseLogger struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (l *responseLogger) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *responseLogger) Write(b []byte) (int, error) {
	n, err := l.ResponseWriter.Write(b)
	l.written += int64(n)
	return n, err
}

// Функция генерации request_id
func generateRequestID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		// В случае ошибки - используем временную метку
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return id.String()
}

type requestContextKey string

const requestIDKey requestContextKey = "requestID"

// LoggingMiddleware создаёт middleware с использованием slog.Logger
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Генерируем ID запроса
			requestID := generateRequestID()

			// Добавляем ID в контекст
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Создаём обёртку для захвата статуса ответа
			loggerWrapper := &responseLogger{ResponseWriter: w}

			// Вызываем следующий обработчик в цепочке
			next.ServeHTTP(loggerWrapper, r)

			duration := time.Since(start)

			// Если статус не был установлен, используем 200
			if loggerWrapper.statusCode == 0 {
				loggerWrapper.statusCode = http.StatusOK
			}

			// Логируем информацию о запросе
			logger.Info("HTTP request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", loggerWrapper.statusCode),
				slog.Int64("bytes_written", loggerWrapper.written),
				slog.Duration("duration", duration),
			)
		})
	}
}
