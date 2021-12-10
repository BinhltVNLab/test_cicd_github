package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

type OvertimeWeight struct {
	cm.BaseModel

	tableName       struct{} `sql:"alias:ow"`
	OrganizationId  int
	NormalDayWeight float64
	WeekendWeight   float64
	HolidayWeight   float64
}
