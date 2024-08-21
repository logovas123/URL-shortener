package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/logger/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

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

//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

// Наш Storage(sqlite) реализует интерфейс URLSaver
// здесь будет возвращаться обработчик, который обрабатывает запрос
func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		// аргументы функции With() будут добавлятся к каждому выводу лога; GetReqID - задёт номер запроса
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
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		var id int64
		alias := req.Alias
		// если алиас пустой, то будем генерировать, избегая ошибки генерации алиаса, который уже существовал
		if alias == "" {
			for {
				alias = random.NewRandomString(aliasLenght)
				id, err = urlSaver.SaveURL(req.URL, alias)
				if errors.Is(err, storage.ErrURLExists) {
					continue
				} else {
					break
				}
			}
		} else {
			// сохраняем url
			id, err = urlSaver.SaveURL(req.URL, alias)
			// обработка ошибки если алиас существует
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.JSON(w, r, resp.Error("url already exists"))

				return
			}
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		// сообщаем об успешном добавлении url
		log.Info("url added", slog.Int64("id", id))

		// возвращаем статус ок
		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
