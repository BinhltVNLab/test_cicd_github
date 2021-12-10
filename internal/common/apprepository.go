package common

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/db"
)

type AppRepository struct {
	DB     *pg.DB
	Logger echo.Logger
}

func (repo *AppRepository) Init(logger echo.Logger) {
	repo.Logger = logger
	repo.DB = db.Init(logger)
}
