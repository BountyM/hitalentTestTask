package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port   string `env:"APP_PORT" envDefault:"8080"`
	DB     DB     `envPrefix:"DB_"`
	Logger Logger `envPrefix:"LOGGER_"`
}

// DB содержит параметры подключения к базе данных
type DB struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"5432"`
	Username string `env:"USERNAME" envDefault:"postgres"`
	Password string `env:"PASSWORD"`
	Dbname   string `env:"NAME" envDefault:"myapp"`
	Sslmode  string `env:"SSLMODE" envDefault:"disable"`
}

// Config для логгера
type Logger struct {
	Level  string `env:"LEVEL" envDefault:"INFO"`
	Format string `env:"FORMAT" envDefault:"json"` // "json" или "text"
}

// Load загружает .env файл из директории internal/config,
// затем парсит переменные окружения в структуру Config.
func Load() (*Config, error) {
	// Определяем путь к текущему файлу (config.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	basepath := filepath.Dir(filename) // internal/config
	envPath := filepath.Join(basepath, ".env")

	// Загружаем .env файл (если есть)
	_ = godotenv.Load(envPath) // игнорируем ошибку - файл может отсутствовать

	// Парсим переменные окружения в структуру Config
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
