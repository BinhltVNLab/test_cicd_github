package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserLeaveRequest : struct for db table user_leave_requests
type UserLeaveRequest struct {
	cm.BaseModel

	tableName            struct{} `sql:"user_leave_requests,alias:ulr"`
	OrganizationID       int
	UserID               int
	LeaveRequestTypeID   int
	DatetimeLeaveFrom    time.Time
	DatetimeLeaveTo      time.Time
	CreatedBy            int
	UpdatedBy            int
	EmailTitle           string
	EmailContent         string
	SubtractDayOffTypeID int
	Reason               string
	Hour                 float64
	CalendarEventId      string
}

type UserLeaveRequestExt struct {
	TableName struct{} `sql:"users,alias:usr"`

	User
	UserProfile      UserProfile        `pg:"fk:user_id"`
	UserLeaveRequest []UserLeaveRequest `pg:"fk:user_id"`
}
