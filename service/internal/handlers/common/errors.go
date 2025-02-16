package common

import (
	"errors"

	"github.com/K1flar/crawlers/internal/business_errors"
)

func ErrorMsg(err error) string {
	switch {
	case errors.Is(err, business_errors.InvalidQuery):
		return "Некорректная тема"
	}

	return "Неизвестная ошибка, повторите позже"
}
