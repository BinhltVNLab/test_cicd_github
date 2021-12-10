package requestparams

import (
	"time"

	"github.com/robfig/cron/v3"
)

// CreateLeaveRequestParams : Param for user leave request
type CreateLeaveRequestParams struct {
	LeaveRequest []LeaveRequest `json:"leave_request"`
}

// LeaveRequest : struct for leave request
type LeaveRequest struct {
	OrgID                int     `json:"organization_id" valid:"required"`
	UserID               int     `json:"user_id" valid:"required"`
	LeaveRequestTypeID   int     `json:"leave_request_type_id" valid:"required"`
	DatetimeLeaveFrom    string  `json:"datetime_leave_from" valid:"required"`
	DatetimeLeaveTo      string  `json:"datetime_leave_to"`
	CreatedBy            int     `json:"created_by" valid:"required"`
	UpdatedBy            int     `json:"updated_by" valid:"required"`
	EmailTitle           string  `json:"email_title" valid:"required"`
	EmailContent         string  `json:"email_content" valid:"required"`
	Reason               string  `json:"reason" valid:"required"`
	Hour                 float64 `json:"hour"`
	SubtractDayOffTypeID int     `json:"subtract_day_off_type_id"`
	ExtraTime            float64 `json:"extra_time"`
}

// CreateLeaveBonusParams : Param for user leave bonus
type CreateLeaveBonusParams struct {
	LeaveBonus []LeaveBonus `json:"leave_bonus"`
}

// LeaveBonus : struct for leave bonus
type LeaveBonus struct {
	OrgID                int     `json:"organization_id"`
	UserID               int     `json:"user_id" valid:"required"`
	LeaveBonusTypeID     int     `json:"leave_bonus_type_id" valid:"required"`
	CreatedBy            int     `json:"created_by"`
	UpdatedBy            int     `json:"updated_by"`
	YearBelong           int     `json:"year_belong" valid:"required"`
	Reason               string  `json:"reason" valid:"required"`
	Hour                 float64 `json:"hour" valid:"required"`
	ExpireBonusLeaveDate string  `json:"expire_bonus_leave_date"`
	HourRemaining        float64 `json:"hour_remaining"`
}

// LeaveStatusParams : Param get day used, day remaining of user
type LeaveStatusParams struct {
	OrgID  int `json:"organization_id" valid:"required"`
	UserID int `json:"user_id" valid:"required"`
	Year   int `json:"year" valid:"required"`
}

// LeaveHistoryParams : Struct for leave history params
type LeaveHistoryParams struct {
	ID                   int      `json:"id"`
	UserID               int      `json:"user_id"`
	UserName             string   `json:"user_name"`
	DatetimeLeaveFrom    string   `json:"datetime_leave_from" valid:"formatDisplayDate"`
	DatetimeLeaveTo      string   `json:"datetime_leave_to" valid:"formatDisplayDate"`
	SubtractDayOffTypeID int      `json:"subtract_day_off_type_id"`
	DateOfWeek           []string `json:"date_of_week"`
	BranchID             int      `json:"branch_id"`
	CurrentPage          int      `json:"current_page"`
	RowPerPage           int
}

// LeaveHistoryRecords : Leave history response
type LeaveHistoryRecords struct {
	UserID               int       `json:"user_id"`
	LeaveRequestTypeID   int       `json:"leave_request_type_id"`
	DatetimeLeaveFrom    time.Time `json:"datetime_leave_from"`
	DatetimeLeaveTo      time.Time `json:"datetime_leave_to"`
	Hour                 float64   `json:"hour"`
	SubtractDayOffTypeID int       `json:"subtract_day_off_type_id"`
	HourTo               int       `json:"hour_to"`
	MinuteTo             int       `json:"minute_to"`
}

// LeaveHistoriesByUser : Leave histories by user response
type LeaveHistoriesByUser struct {
	UserID        int                  `json:"user_id"`
	HistoriesUser []LeaveHistoryByUser `json:"histories"`
}

// LeaveHistoryByUser : Leave history by user response
type LeaveHistoryByUser struct {
	LeaveRequestTypeID   int      `json:"leave_request_type_id"`
	SubtractDayOffTypeID int      `json:"subtract_day_off_type_id"`
	LeaveDates           []string `json:"leave_dates"`
}

// RemoveLeaveParams : Remove leave request params
type RemoveLeaveParams struct {
	LeaveID int `json:"leave_id" valid:"required"`
}

// RemoveCronLeaveParam : Remove cron annual leave bonus param
type RemoveCronLeaveParam struct {
	ID cron.EntryID `json:"id" valid:"required"`
}

// LeaveRequestListParams : Leave request list params
type LeaveRequestListParams struct {
	UserName           string `json:"user_name"`
	LeaveRequestTypeID int    `json:"leave_request_type_id"`
	Branch             int    `json:"branch"`
	DatetimeLeaveFrom  string `json:"datetime_leave_from"`
	DatetimeLeaveTo    string `json:"datetime_leave_to"`
	CurrentPage        int    `json:"current_page"`
	RowPerPage         int    `json:"row_per_page"`
}

// LeaveRequestRecords : Leave request records
type LeaveRequestRecords struct {
	ID                   int       `json:"id"`
	UserID               int       `json:"user_id"`
	LeaveRequestTypeID   int       `json:"leave_request_type_id"`
	Email                string    `json:"email"`
	Avatar               string    `json:"avatar"`
	FullName             string    `json:"full_name"`
	DatetimeLeaveFrom    time.Time `json:"datetime_leave_from"`
	DatetimeLeaveTo      time.Time `json:"datetime_leave_to"`
	EmailContent         string    `json:"email_content"`
	EmailTitle           string    `json:"email_title"`
	Reason               string    `json:"reason"`
	HourFrom             int       `json:"hour_from"`
	MinuteFrom           int       `json:"minute_from"`
	HourTo               int       `json:"hour_to"`
	MinuteTo             int       `json:"minute_to"`
	SubtractDayOffTypeID int       `json:"subtract_day_off_type_id"`
}

type GetLeaveBonusParam struct {
	FullName         string `json:"full_name"`
	LeaveBonusTypeId int    `json:"leave_bonus_type_id"`
	Year             int    `json:"year"`
	CurrentPage      int    `json:"current_page" valid:"required"`
	RowPerPage       int    `json:"row_per_page" valid:"required"`
	IsDeleted        bool   `json:"is_deleted"`
}

type LeaveBonusRecords struct {
	Id               int       `json:"id"`
	UserId           int       `json:"user_id"`
	CreatedBy        string    `json:"created_by"`
	Reason           string    `json:"reason"`
	Year             int       `json:"year"`
	Hour             float64   `json:"hour"`
	LeaveBonusTypeId int       `json:"leave_bonus_type_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type ValidLeaveBonusRecords struct {
	Id            int     `json:"id"`
	Hour          float64 `json:"hour"`
	HourRemaining float64 `json:"hour_remaining"`
}

type ExpireLeaveBonusRecords struct {
	UserId    int     `json:"user_id"`
	TotalHour float64 `json:"total_hour"`
}

type StartDateLeaveRecords struct {
	Id               int       `json:"id"`
	UpdatedAt        time.Time `json:"updated_at"`
	Hour             float64   `json:"hour"`
	LeaveBonusTypeId int       `json:"leave_bonus_type_id"`
}

type RemoveLeaveBonusParam struct {
	Id        int  `json:"id" valid:"required"`
	IsDeleted bool `json:"is_deleted"`
}

type GetLeaveBonusByIdParam struct {
	Id int `json:"id" valid:"required"`
}

type EditLeaveBonusParam struct {
	Id               int     `json:"id" valid:"required"`
	LeaveBonusTypeId int     `json:"leave_bonus_type_id"`
	YearBelong       int     `json:"year_belong"`
	Hour             float64 `json:"hour"`
	Reason           string  `json:"reason"`
}

type UpdateHourRemainingLeaveParam struct {
	OrgID     int     `json:"organization_id" valid:"required"`
	UserId    int     `json:"user_id" valid:"required"`
	HourLeave float64 `json:"hour_leave" valid:"required"`
}
