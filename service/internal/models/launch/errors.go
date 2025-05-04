package launch

import "github.com/K1flar/crawlers/internal/utils"

const (
	UnknownErrorSlug ErrorSlug = "unknown"
)

func ErrorToSlug(err error) *ErrorSlug {
	if err == nil {
		return nil
	}

	switch {
	// TODO: замапить ошибки
	}

	return utils.Ptr(UnknownErrorSlug)
}
