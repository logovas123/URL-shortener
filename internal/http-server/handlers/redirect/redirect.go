package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/logger/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// интерфейс для получения url по алиасу
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

// возвращает обработчик который возвращает url (GetURL)
func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		// получаем параметр alias из роутера, т.е. /{alias}
		alias := chi.URLParam(r, "alias")
		// если алиас пришёл пустым
		if alias == "" {
			log.Info("alias not empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		// если алиас не пустой, то получаем url
		resURL, err := urlGetter.GetURL(alias)
		// если алиас не найден
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		// обрабатываем общую ошибку
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}
		// сообщаем что url получен
		log.Info("got url", slog.String("url", resURL))

		// redirect to found url
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
