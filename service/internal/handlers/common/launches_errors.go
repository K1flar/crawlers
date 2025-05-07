package common

import (
	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/utils"
)

var errorSlugToMsg = map[launch.ErrorSlug]string{
	launch.SearxErrorSlug:       "Ошибка поисковой системы SearX, попробуйте позже",
	launch.ZeroStartSourcesSlug: "Не нашли стартовые источники для запуска робота, попробуйте перезапустить",
}

func ErrorSlugToMsg(slug *launch.ErrorSlug) *string {
	if slug == nil {
		return nil
	}

	if msg, ok := errorSlugToMsg[*slug]; ok {
		return &msg
	}

	return utils.Ptr("Неизвестная ошибка, попробуйте позже")
}
