package common

import (
	"encoding/json"
	"io"
	"net/http"
)

func DTO[T any](r *http.Request) (T, error) {
	var dto T

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return dto, err
	}
	defer r.Body.Close()

	if err = json.Unmarshal(b, &dto); err != nil {
		return dto, err
	}

	return dto, nil
}
