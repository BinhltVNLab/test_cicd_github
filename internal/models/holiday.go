package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	"time"
)

type Holiday struct {
	cm.BaseModel

	tableName      struct{} `sql:"alias:hld"`
	OrganizationId int
	HolidayDate    time.Time
	Description    string
}
