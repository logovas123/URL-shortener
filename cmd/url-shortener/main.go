package main

import (
	"fmt"
	"log/slog"
	"os"

	"url-shortener/internal/config"
	"url-shortener/internal/http-server/middleware/mwLogger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)

	log := setupLogger(cfg.Env)

	// какая то инфа с логера, и выводим также окружение, которое используется
	// если меняем переменную, то формат вывода меняется согласно настройкам (json например для envDev)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// создаём хранилище
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		// Err() вернёт ключ-значение для вывода ошибки
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	// создали новый роутер
	router := chi.NewRouter()

	// middleware
	// для каждого запроса будет свой id, чтобы проще отлавливать ошибку
	router.Use(middleware.RequestID)
	// логгер из коробки chi
	router.Use(middleware.Logger)
	// свой логер
	router.Use(mwLogger.New(log))
	// если паника внутри хендлера, то будем востанавливать эту панику, чтобы приложение не упало
	router.Use(middleware.Recoverer)
	// чтобы urlы были красивыми
	router.Use(middleware.URLFormat)

	// TODO: run server
}

// функция возвращает логер, который будет зависеть от env, то есть для каждой среды окружения(prod, local, dev) свой логер
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		// возвращаем "красивый" логер
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
