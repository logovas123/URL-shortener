package deleter

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// интерфейс для удаления url
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
	GetURL(alias string) (string, error)
}

// возвращает обработчик который удаляет url
func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		// получаем параметр alias из роутера, т.е. /{alias}
		alias := chi.URLParam(r, "alias")
		fmt.Print(alias)
		// если алиас пришёл пустым
		if alias == "" {
			log.Info("alias not empty")

			return
		}

		// если алиас не пустой, то получаем url
		resURL, err := urlDeleter.GetURL(alias)
		// если алиас не найден
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			return
		}
		// обрабатываем общую ошибку
		if err != nil {
			log.Error("failed to find url", sl.Err(err))

			return
		}

		err = urlDeleter.DeleteURL(alias)
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))

			return
		}

		// сообщаем что url удален
		log.Info("successfully delete url", slog.String("url", resURL))

		// http.Redirect(w, r, resURL, http.StatusFound)
	}
}
