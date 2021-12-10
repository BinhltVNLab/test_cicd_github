package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	"time"
)

// UserLeaveBonus : struct for db table user_leave_bonus
type UserOvertimeRequest struct {
	cm.BaseModel

	tableName            struct{} `sql:"alias:uotr"`
	UserId               int
	ProjectId            int
	Status               int
	DatetimeOvertimeFrom time.Time
	DatetimeOvertimeTo   time.Time
	EmailTitle           string
	EmailContent         string
	Reason               string
	OvertimeType         int
	WorkAtNoon           int
	SendTo               []string `pg:",array"`
	SendCc               []string `pg:",array"`
}
