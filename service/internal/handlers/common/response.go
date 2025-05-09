package common

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/K1flar/crawlers/internal/business_errors"
)

func OK(w http.ResponseWriter, body any) {
	w.WriteHeader(http.StatusOK)

	b, err := json.Marshal(body)
	if err != nil {
		b = []byte("{}")
	}

	w.Write(b)
}

func BadRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)

	b, err := json.Marshal(map[string]string{
		"error": msg,
	})
	if err != nil {
		b = []byte(`{"error": "bad request"}`)
	}

	w.Write(b)
}

func Forbidden(w http.ResponseWriter, code string, msg string) {
	w.WriteHeader(http.StatusForbidden)

	b, err := json.Marshal(map[string]string{
		"code":  code,
		"error": msg,
	})
	if err != nil {
		b = []byte(`{"error": "forbidden"}`)
	}

	w.Write(b)
}

func InternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error": "internal error"}`))
}

func Error(w http.ResponseWriter, err error) {
	var businessError *business_errors.BusinessError
	if errors.As(err, &businessError) {
		Forbidden(w, businessError.Code, ErrorMsg(err))
		return
	}

	InternalError(w)
}
