package services

import (
	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/task"
)

type LaunhToFinishParams struct {
	LaunchID int64
	Task     task.Task
	Pages    map[string]*page.PageWithParentURL
	Error    error
}
