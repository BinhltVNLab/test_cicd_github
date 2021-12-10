package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserLeaveBonus : struct for db table user_leave_bonus
type UserLeaveBonus struct {
	cm.BaseModel

	tableName            struct{} `sql:"alias:ulb"`
	OrganizationID       int
	UserID               int
	LeaveBonusTypeID     int
	CreatedBy            int
	UpdatedBy            int
	YearBelong           int
	Reason               string
	Hour                 float64
	ExpireBonusLeaveDate string
	HourRemaining        float64
}
