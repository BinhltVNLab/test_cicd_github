package requestparams

import "time"

// SeachTimekeepingUserParams : Params for seach all timekeeping of user
type SeachTimekeepingUserParams struct {
	DateFrom    string `json:"date_from"`
	DateTo      string `json:"date_to"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// SeachAllTimekeepingParams : Params for seach all timekeeping
type SeachAllTimekeepingParams struct {
	FromDate    string `json:"date_from"`
	ToDate      string `json:"date_to"`
	UserName    string `json:"user_name"`
	BranchID    int    `json:"branch_id"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// UserTimekeepingResponse : struct for user timekeeping response
type UserTimekeepingResponse struct {
	tableName      struct{} `pg:",discard_unknown_columns"`
	OrganizationID int
	UserID         int
	UserName       string
	Email          string
	Branch         int
	CheckInTime    time.Time
	CheckOutTime   time.Time
}

// AllUserTimekeepingResponse : struct for all user timekeeping response
type AllUserTimekeepingResponse struct {
	tableName struct{} `pg:",discard_unknown_columns"`

	MaxCheckInID  int
	MaxCheckOutID int
	CheckInTime   time.Time
	CheckOutTime  time.Time
}

type TkExportExcelParams struct {
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type TkExportExcelRecords struct {
	FullName     string    `json:"full_name"`
	Date         time.Time `json:"date"`
	CheckInTime  time.Time `json:"check_in_time"`
	CheckOutTime time.Time `json:"check_out_time"`
}
