package common

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	db "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/db"
)

type BaseController struct {
	DB     *pg.DB
	Logger echo.Logger
}

func (repo *BaseController) Init(logger echo.Logger) {
	repo.Logger = logger
	repo.DB = db.Init(logger)
}
