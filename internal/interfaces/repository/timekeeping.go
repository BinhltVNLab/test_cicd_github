package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// TimekeepingRepository : Timekeeping Repository
type TimekeepingRepository interface {
	InsertCheckInTime(orgID int, userID int) error
	InsertCheckOutTime(ID int) error
	GetTimekeepingWithMaxCheckInTime(orgID int, userID int) (m.UserTimekeeping, error)
	GetLastTimekeepingToday(orgID int, userID int) (m.UserTimekeeping, error)
	GetAllTimekeepingUser(orgID int, userID int, seachTimekeepingUserParams *param.SeachTimekeepingUserParams) ([]param.UserTimekeepingResponse, int, error)
	GetAllTimekeeping(orgID int, seachAllTimekeepingParams *param.SeachAllTimekeepingParams) ([]param.UserTimekeepingResponse, int, error)
	SelectTimekeepingsByDate(organizationId int, exportCSVParams *param.TkExportExcelParams) ([]param.TkExportExcelRecords, error)
}
