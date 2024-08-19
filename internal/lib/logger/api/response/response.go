package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Response будет использоваться в разных хендлерах, поэтому его объявляем здесь
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

// в массив получаем список ошибок валидатора
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	// перебираем список ошибок
	for _, err := range errs {
		switch err.ActualTag() { // смотрим на тег ошибки
		case "required": // то есть суть в том что если, например, стоит данный тег мы в ошибке сообщаем что какое то поле было обязательным, то есть делаем ошибки человекочитаемыми
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
