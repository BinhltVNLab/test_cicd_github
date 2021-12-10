package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// LeaveRepository interface
type LeaveRepository interface {
	InsertLeaveRequest(
		leaveRequestParams *param.LeaveRequest,
		holidayRepo HolidayRepository,
		userRepo UserRepository,
		notificationRepo NotificationRepository,
		uniqueUsersId []int,
	) (int, string, string, float64, error)
	InsertLeaveBonus(leaveBonus *param.LeaveBonus) error
	InsertLeaveBonusOvertimeWithTx(tx *pg.Tx, leaveBonusParams *param.LeaveBonus) error
	InsertLeaveBonusWithTx(organizationId int, createdBy int, leaveBonusParams *[]param.LeaveBonus) error
	UpdateLeaveRequest(Id int, calendarEventId string) error
	CountHourUsed(orgID int, userID int, year int) (float64, error)
	CountHourBonus(orgID int, userID int, year int) (float64, error)
	CountHoursUpToExpirationDate(orgID int, userID int, date string, firstDate string) (float64, error)
	LeaveHistory(orgID int, leaveHistoryParams *param.LeaveHistoryParams) ([]param.LeaveHistoryRecords, error)
	LeaveRequests(
		organizationID int,
		orderBy string,
		leaveRequestListParams *param.LeaveRequestListParams,
		isPagination bool,
		columns ...string,
	) ([]param.LeaveRequestRecords, int, error)
	RemoveLeave(leaveID int) error
	GetLeaveDayStatus(orgID int, userID int, year int) (float64, float64, float64, float64)
	SelectLeaveRequestById(Id int) (m.UserLeaveRequest, error)
	SelectLeaveBonuses(orgID int, params *param.GetLeaveBonusParam) ([]param.LeaveBonusRecords, int, error)
	UpdateDeleted(id int, isDelete bool) error
	SelectLeaveBonusById(Id int) (m.UserLeaveBonus, error)
	UpdateLeaveBonus(userId int, params *param.EditLeaveBonusParam) error
	SearchUserLeave(orgID int, leaveHistoryParams *param.LeaveHistoryParams) ([]m.UserLeaveRequestExt, int, error)
	ClearExpireDate(orgID int) error
	UpdateHourRemainingLeaveByID(id int, hourRemaining float64) error
	SelectValidDateLeave(orgID int, userID int) ([]param.ValidLeaveBonusRecords, error)
	SelectValidDateOfUser(orgID int, userID int) ([]param.ValidLeaveBonusRecords, error)
	SelectStartLeaveCurrentYear(orgID int) ([]param.StartDateLeaveRecords, error)
	UpdateExpireLeaveDate(id int, hour float64, expireLeaveDate string) error
	CountExpireDate(orgID int) ([]param.ExpireLeaveBonusRecords, error)
}
