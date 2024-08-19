package slogdiscard

import (
	"context"
	"log/slog"
)

// в slogdiscard будет логер для тестов, который будет удалять все свои сообщения, которые касаются тестов для хендлера
// возвращает логер
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

// хендлер который выкидывает ссообщения
type DiscardHandler struct{}

// билдер для хендлера, который выбрасывает сообщения
func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// Просто игнорируем запись журнала
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Возвращает тот же обработчик, так как нет атрибутов для сохранения
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	// Возвращает тот же обработчик, так как нет группы для сохранения
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// Всегда возвращает false, так как запись журнала игнорируется
	return false
}
