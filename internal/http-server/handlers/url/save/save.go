package save

import (
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/logger/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// к нам будет поступать запрос, к котором будет находиться json объект, который описывает url, который нужно сохранить
// validate говорит говорит validator'у что поле URL обязательное, а также валидатор будет определять действительно url лежит в этом поле
type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

// ответ от сервиса
type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"` // Alias возвращаем, потому что в запросе он будет необязательным параметром, если в запросе его не будет мы будем его генерировать
}

const aliasLenght = 6

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

// Наш Storage(sqlite) реализует интерфейс URLSaver
// здесь будет возвращаться обработчик, который обрабатывает запрос
func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		// логер возвращает инфу об id запроса
		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request

		// декодируем тело запроса в json струтктуру
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			// пишем ошибку в лог
			log.Error("failed to decode request body", sl.Err(err))

			// возвращаем json с ответом клиенту, если ошибка(в виде html)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		// сообщаем об успешном декодировании
		log.Info("request body decoded", slog.Any("request", req))

		// проверяем на валидацию данных структуру (json - запрос), если получена ошибка, то
		// получаем список ошибок и функция ValidationError(), вернёт информацию по каждой ошибке на понятном языке
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			// возвращаем json с ответом клиенту, если ошибка(в виде html)
			render.JSON(w, r, resp.Error("invalid request"))
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLenght)
		}
	}
}
