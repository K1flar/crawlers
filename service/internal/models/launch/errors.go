package launch

import (
	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/utils"
)

const (
	SearxErrorSlug       ErrorSlug = "searx_error"
	ZeroStartSourcesSlug ErrorSlug = "zero_start_sources"
	UnknownErrorSlug     ErrorSlug = "unknown"
)

var errorToSlug = map[error]ErrorSlug{
	business_errors.SearxError:       SearxErrorSlug,
	business_errors.ZeroStartSources: ZeroStartSourcesSlug,
}

func ErrorToSlug(err error) *ErrorSlug {
	if err == nil {
		return nil
	}

	if slug, ok := errorToSlug[err]; ok {
		return &slug
	}

	return utils.Ptr(UnknownErrorSlug)
}
