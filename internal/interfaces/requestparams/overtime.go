package requestparams

import "time"

type CreateOvertimeParams struct {
	UserId               int      `json:"user_id" valid:"required"`
	ProjectId            int      `json:"project_id" valid:"required"`
	Status               int      `json:"status" valid:"required,range(1|3)"`
	DatetimeOvertimeFrom string   `json:"datetime_overtime_from" valid:"required"`
	DatetimeOvertimeTo   string   `json:"datetime_overtime_to" valid:"required"`
	EmailTitle           string   `json:"email_title"`
	EmailContent         string   `json:"email_content"`
	Reason               string   `json:"reason"`
	OvertimeType         int      `json:"overtime_type" valid:"required"`
	WorkAtNoon           int      `json:"work_at_noon" valid:"required"`
	SendTo               []string `json:"send_to"`
	SendCc               []string `json:"send_cc"`
	UsersIdNotification  []int    `json:"users_id_notification"`
}

type GetOvertimeRequestsParams struct {
	Id           int    `json:"id"`
	UsersId      []int  `json:"users_id"`
	ProjectId    int    `json:"project_id"`
	Status       int    `json:"status"`
	Branch       int    `json:"branch"`
	DateFrom     string `json:"date_from"`
	DateTo       string `json:"date_to"`
	OvertimeType int    `json:"overtime_type"`
	CurrentPage  int    `json:"current_page"`
	RowPerPage   int    `json:"row_per_page"`
}

type OvertimeRequestsRecords struct {
	Id                   int       `json:"id"`
	EmployeeId           string    `json:"employee_id"`
	FullName             string    `json:"full_name"`
	Branch               string    `json:"branch"`
	ProjectName          string    `json:"project_name"`
	Status               int       `json:"status"`
	OvertimeType         int       `json:"overtime_type"`
	WorkAtNoon           int       `json:"work_at_noon"`
	DatetimeOvertimeFrom time.Time `json:"datetime_overtime_from"`
	DatetimeOvertimeTo   time.Time `json:"datetime_overtime_to"`
	HourFrom             int       `json:"hour_from"`
	MinuteFrom           int       `json:"minute_from"`
	HourTo               int       `json:"hour_to"`
	MinuteTo             int       `json:"minute_to"`
}

type GetOvertimeRequestParam struct {
	Id int `json:"id" valid:"required"`
}

type UpdateOvertimeRequestParams struct {
	Id                   int      `json:"id" valid:"required"`
	UserId               int      `json:"user_id" valid:"required"`
	ProjectId            int      `json:"project_id"`
	DatetimeOvertimeFrom string   `json:"datetime_overtime_from"`
	DatetimeOvertimeTo   string   `json:"datetime_overtime_to"`
	EmailTitle           string   `json:"email_title"`
	EmailContent         string   `json:"email_content"`
	Reason               string   `json:"reason"`
	OvertimeType         int      `json:"overtime_type"`
	SendTo               []string `json:"send_to"`
	SendCc               []string `json:"send_cc"`
}

type CreateOvertimeWeightParams struct {
	NormalDayWeight float64 `json:"normal_day_weight" valid:"required"`
	WeekendWeight   float64 `json:"weekend_weight" valid:"required"`
	HolidayWeight   float64 `json:"holiday_weight" valid:"required"`
}

type EditOvertimeWeightParams struct {
	Id              int     `json:"id" valid:"required"`
	NormalDayWeight float64 `json:"normal_day_weight"`
	WeekendWeight   float64 `json:"weekend_weight"`
	HolidayWeight   float64 `json:"holiday_weight"`
}
