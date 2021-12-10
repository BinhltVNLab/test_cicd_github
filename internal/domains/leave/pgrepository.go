package leave

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

// PgLeaveRepository : Struct repository
type PgLeaveRepository struct {
	cm.AppRepository
}

// NewPgLeaveRepository : Init repository
func NewPgLeaveRepository(logger echo.Logger) (repo *PgLeaveRepository) {
	repo = &PgLeaveRepository{}
	repo.Init(logger)
	return
}

// InsertLeaveRequest : Insert leave request to database
func (repo *PgLeaveRepository) InsertLeaveRequest(
	leaveRequestParams *param.LeaveRequest,
	holidayRepo rp.HolidayRepository,
	userRepo rp.UserRepository,
	notificationRepo rp.NotificationRepository,
	uniqueUsersId []int,
) (int, string, string, float64, error) {
	var id int
	var content string
	var link string
	var hour float64

	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		datetimeLeaveFrom := calendar.ParseTime(cf.FormatDateNoSec, leaveRequestParams.DatetimeLeaveFrom)
		datetimeLeaveTo := calendar.ParseTime(cf.FormatDateNoSec, leaveRequestParams.DatetimeLeaveTo)
		leaveRequest := m.UserLeaveRequest{
			OrganizationID:       leaveRequestParams.OrgID,
			UserID:               leaveRequestParams.UserID,
			LeaveRequestTypeID:   leaveRequestParams.LeaveRequestTypeID,
			DatetimeLeaveFrom:    datetimeLeaveFrom,
			DatetimeLeaveTo:      datetimeLeaveTo,
			CreatedBy:            leaveRequestParams.CreatedBy,
			UpdatedBy:            leaveRequestParams.UpdatedBy,
			EmailTitle:           leaveRequestParams.EmailTitle,
			EmailContent:         leaveRequestParams.EmailContent,
			SubtractDayOffTypeID: leaveRequestParams.SubtractDayOffTypeID,
			Reason:               leaveRequestParams.Reason,
			Hour: calendar.CalculateHour(
				leaveRequestParams.OrgID,
				holidayRepo,
				leaveRequestParams.LeaveRequestTypeID,
				datetimeLeaveFrom,
				datetimeLeaveTo,
				leaveRequestParams.SubtractDayOffTypeID,
				leaveRequestParams.ExtraTime,
			),
		}
		transErr = tx.Insert(&leaveRequest)
		if transErr != nil {
			return transErr
		}
		id = leaveRequest.ID
		hour = leaveRequest.Hour

		notificationParams := new(param.InsertNotificationParam)
		notificationParams.Content = "has just created a leave request"
		notificationParams.RedirectUrl = "/hrm/leave/history-user-leave?id=" + strconv.Itoa(id) +
			"&user_id=" + strconv.Itoa(leaveRequestParams.UserID) +
			"&date_from=" + strings.Split(leaveRequestParams.DatetimeLeaveFrom, " ")[0] +
			"&date_to=" + strings.Split(leaveRequestParams.DatetimeLeaveTo, " ")[0]

		for _, userId := range uniqueUsersId {
			if userId == leaveRequest.UserID {
				continue
			}
			notificationParams.Receiver = userId

			transErr = notificationRepo.InsertNotificationWithTx(tx, leaveRequestParams.OrgID, leaveRequestParams.UserID, notificationParams)
			if transErr != nil {
				return transErr
			}
		}

		fullName, transErr := userRepo.SelectFullNameUser(leaveRequestParams.UserID)
		if transErr != nil && transErr.Error() != pg.ErrNoRows.Error() {
			return transErr
		}

		content = fullName + " " + notificationParams.Content
		link = notificationParams.RedirectUrl

		return transErr
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return id, content, link, hour, err
}

// InsertLeaveBonus : Insert leave bonus to database
func (repo *PgLeaveRepository) InsertLeaveBonus(leaveBonusParams *param.LeaveBonus) error {
	leaveBonus := m.UserLeaveBonus{
		OrganizationID:   leaveBonusParams.OrgID,
		UserID:           leaveBonusParams.UserID,
		LeaveBonusTypeID: leaveBonusParams.LeaveBonusTypeID,
		CreatedBy:        leaveBonusParams.CreatedBy,
		UpdatedBy:        leaveBonusParams.UpdatedBy,
		YearBelong:       leaveBonusParams.YearBelong,
		Reason:           leaveBonusParams.Reason,
		Hour:             leaveBonusParams.Hour,
	}

	err := repo.DB.Insert(&leaveBonus)

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// InsertLeaveBonus : Insert leave bonus to database
func (repo *PgLeaveRepository) InsertLeaveBonusOvertimeWithTx(tx *pg.Tx, leaveBonusParams *param.LeaveBonus) error {
	leaveBonus := m.UserLeaveBonus{
		OrganizationID:       leaveBonusParams.OrgID,
		UserID:               leaveBonusParams.UserID,
		LeaveBonusTypeID:     leaveBonusParams.LeaveBonusTypeID,
		CreatedBy:            leaveBonusParams.CreatedBy,
		UpdatedBy:            leaveBonusParams.UpdatedBy,
		YearBelong:           leaveBonusParams.YearBelong,
		Reason:               leaveBonusParams.Reason,
		Hour:                 leaveBonusParams.Hour,
		ExpireBonusLeaveDate: leaveBonusParams.ExpireBonusLeaveDate,
		HourRemaining:        leaveBonusParams.HourRemaining,
	}

	err := tx.Insert(&leaveBonus)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) InsertLeaveBonusWithTx(organizationId int, createdBy int, leaveBonusParams *[]param.LeaveBonus) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		for _, bonusParam := range *leaveBonusParams {
			leaveBonus := m.UserLeaveBonus{
				OrganizationID:       organizationId,
				UserID:               bonusParam.UserID,
				LeaveBonusTypeID:     bonusParam.LeaveBonusTypeID,
				CreatedBy:            createdBy,
				UpdatedBy:            createdBy,
				YearBelong:           bonusParam.YearBelong,
				Reason:               bonusParam.Reason,
				Hour:                 bonusParam.Hour,
				HourRemaining:        bonusParam.HourRemaining,
				ExpireBonusLeaveDate: bonusParam.ExpireBonusLeaveDate,
			}

			errTx = tx.Insert(&leaveBonus)
			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}
		}

		return errTx
	})

	return err
}

// CountHourUsed : Sum hour used of user
func (repo *PgLeaveRepository) CountHourUsed(orgID int, userID int, year int) (float64, error) {
	var hour float64
	err := repo.DB.Model(&m.UserLeaveRequest{}).
		ColumnExpr("SUM(hour)").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("date_part('year', datetime_leave_from) = ?", year).
		Select(&hour)

	return hour, err
}

// CountHourBonus : Sum hour remaining of user
func (repo *PgLeaveRepository) CountHourBonus(orgID int, userID int, year int) (float64, error) {
	var hour float64
	err := repo.DB.Model(&m.UserLeaveBonus{}).
		ColumnExpr("SUM(hour)").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("year_belong = ?", year).
		Select(&hour)

	return hour, err
}

// CountHourRemaining : Sum hour remaining of user
func (repo *PgLeaveRepository) CountHourRemaining(orgID int, userID int, year int) (float64, error) {
	var hour float64
	err := repo.DB.Model(&m.UserLeaveBonus{}).
		ColumnExpr("SUM(hour_remaining)").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("year_belong = ?", year).
		Select(&hour)

	return hour, err
}

// CountHoursUpToExpirationDate : Number of hours up to the expiration date
func (repo *PgLeaveRepository) CountHoursUpToExpirationDate(orgID int, userID int, date string, firstDate string) (float64, error) {
	var hour float64
	err := repo.DB.Model(&m.UserLeaveRequest{}).
		ColumnExpr("SUM(hour)").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("date(ulr.datetime_leave_from) <= DATE(?)", date).
		Where("date(ulr.datetime_leave_from) >= DATE(?)", firstDate).
		Select(&hour)

	return hour, err
}

// LeaveHistory : Select leave history of user
func (repo *PgLeaveRepository) LeaveHistory(orgID int, leaveHistoryParams *param.LeaveHistoryParams) ([]param.LeaveHistoryRecords, error) {
	var leaveRequest []m.UserLeaveRequest
	var leaveHistoryResponse []param.LeaveHistoryRecords

	queryObj := repo.DB.Model(&leaveRequest)
	queryObj.Column("ulr.user_id")
	queryObj.Column("leave_request_type_id", "datetime_leave_from", "datetime_leave_to", "subtract_day_off_type_id", "hour")
	queryObj.ColumnExpr("EXTRACT(HOUR FROM datetime_leave_to) AS hour_to")
	queryObj.ColumnExpr("EXTRACT(MINUTE FROM datetime_leave_to) AS minute_to")
	queryObj.Join("join user_profiles as usr on usr.user_id = ulr.user_id")
	queryObj.Where("organization_id = ?", orgID)

	if leaveHistoryParams.ID != 0 {
		queryObj.Where("id = ?", leaveHistoryParams.ID)
	}

	if leaveHistoryParams.UserName != "" {
		userName := "%" + leaveHistoryParams.UserName + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(usr.first_name)) || ' ' || vietnamese_unaccent(LOWER(usr.last_name)) LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if leaveHistoryParams.UserID != 0 {
		queryObj.Where("ulr.user_id = ?", leaveHistoryParams.UserID)
	}

	if leaveHistoryParams.DatetimeLeaveFrom != "" {
		queryObj.Where("date(datetime_leave_from) >= DATE(?)", leaveHistoryParams.DatetimeLeaveFrom)
	}

	if leaveHistoryParams.DatetimeLeaveTo != "" {
		queryObj.Where("date(datetime_leave_from) <= DATE(?)", leaveHistoryParams.DatetimeLeaveTo)
	}

	if leaveHistoryParams.SubtractDayOffTypeID != 0 {
		queryObj.Where("subtract_day_off_type_id = ?", leaveHistoryParams.SubtractDayOffTypeID)
	}

	queryObj.Order("ulr.user_id ASC")
	err := queryObj.Select(&leaveHistoryResponse)

	return leaveHistoryResponse, err
}

// LeaveRequests : Select leave requests of user
func (repo *PgLeaveRepository) LeaveRequests(
	organizationID int,
	orderBy string,
	leaveRequestListParams *param.LeaveRequestListParams,
	isPagination bool,
	columns ...string,
) ([]param.LeaveRequestRecords, int, error) {
	var leaveRequest []m.UserLeaveRequest
	var leaveRequestRecords []param.LeaveRequestRecords

	queryObj := repo.DB.Model(&leaveRequest)
	queryObj.Column(columns...)
	if !isPagination {
		queryObj.ColumnExpr("EXTRACT(HOUR FROM datetime_leave_from) AS hour_from")
		queryObj.ColumnExpr("EXTRACT(MINUTE FROM datetime_leave_from) AS minute_from")
	}
	queryObj.ColumnExpr("up.first_name || ' ' || up.last_name full_name")
	queryObj.ColumnExpr("EXTRACT(HOUR FROM datetime_leave_to) AS hour_to")
	queryObj.ColumnExpr("EXTRACT(MINUTE FROM datetime_leave_to) AS minute_to")
	queryObj.Join("JOIN user_profiles AS up ON up.user_id = ulr.user_id")
	queryObj.Join("JOIN users AS u ON u.id = ulr.user_id")
	queryObj.Where("ulr.organization_id = ?", organizationID)

	if leaveRequestListParams.UserName != "" {
		userName := "%" + leaveRequestListParams.UserName + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if leaveRequestListParams.LeaveRequestTypeID != 0 {
		queryObj.Where("ulr.leave_request_type_id = ?", leaveRequestListParams.LeaveRequestTypeID)
	}

	if leaveRequestListParams.Branch != 0 {
		queryObj.Where("up.branch = ?", leaveRequestListParams.Branch)
	}

	if leaveRequestListParams.DatetimeLeaveFrom != "" {
		queryObj.Where("date(ulr.datetime_leave_from) >= to_date(?,'YYYY-MM-DD')", leaveRequestListParams.DatetimeLeaveFrom)
	}

	if leaveRequestListParams.DatetimeLeaveTo != "" {
		queryObj.Where("date(ulr.datetime_leave_from) <= to_date(?,'YYYY-MM-DD')", leaveRequestListParams.DatetimeLeaveTo)
	}

	queryObj.OrderExpr(orderBy)
	if isPagination {
		queryObj.Offset((leaveRequestListParams.CurrentPage - 1) * leaveRequestListParams.RowPerPage)
		queryObj.Limit(leaveRequestListParams.RowPerPage)
	}
	totalRow, err := queryObj.SelectAndCount(&leaveRequestRecords)
	if err != nil {
		repo.Logger.Error(err)
	}

	return leaveRequestRecords, totalRow, err
}

func (repo *PgLeaveRepository) UpdateLeaveRequest(Id int, calendarEventId string) error {
	_, err := repo.DB.Model(&m.UserLeaveRequest{CalendarEventId: calendarEventId}).
		Column("calendar_event_id", "updated_at").
		Where("id = ?", Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) SelectLeaveRequestById(Id int) (m.UserLeaveRequest, error) {
	var leaveRequest m.UserLeaveRequest
	err := repo.DB.Model(&leaveRequest).
		Column("calendar_event_id", "organization_id", "user_id", "hour").
		Where("id = ?", Id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return leaveRequest, err
}

// RemoveLeave : Remove leave request
func (repo *PgLeaveRepository) RemoveLeave(leaveID int) error {
	_, err := repo.DB.Model(&m.UserLeaveRequest{}).
		Where("id = ?", leaveID).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) GetLeaveDayStatus(orgID int, userID int, year int) (float64, float64, float64, float64) {
	hourUsedPrevious, err := repo.CountHourUsed(orgID, userID, year-1)
	if err != nil {
		panic(err)
	}

	hourBonusPrevious, err := repo.CountHourBonus(orgID, userID, year-1)
	if err != nil {
		panic(err)
	}

	hourUsed, err := repo.CountHourUsed(orgID, userID, year)
	if err != nil {
		panic(err)
	}

	hourBonus, err := repo.CountHourBonus(orgID, userID, year)
	if err != nil {
		panic(err)
	}

	hourRemaining, err := repo.CountHourRemaining(orgID, userID, year)
	if err != nil {
		panic(err)
	}

	dayRemainingPrevious := (hourBonusPrevious - hourUsedPrevious) / 8
	if dayRemainingPrevious == -0 {
		dayRemainingPrevious = 0
	}

	dayUsed, dayBonus, dayRemaining := hourUsed/8, hourBonus/8, hourRemaining/8

	return dayUsed, dayBonus, dayRemaining, dayRemainingPrevious
}

func (repo *PgLeaveRepository) SelectLeaveBonuses(orgID int, params *param.GetLeaveBonusParam) ([]param.LeaveBonusRecords, int, error) {
	var totalRow int
	var records []param.LeaveBonusRecords
	var condition string

	q1 := "SELECT COUNT(*) FROM user_leave_bonus AS ulb " +
		"JOIN user_profiles AS up ON up.user_id = ulb.user_id " +
		"JOIN user_profiles AS up2 ON up2.user_id = ulb.created_by "

	q2 := "SELECT ulb.id, ulb.user_id, ulb.reason, ulb.hour, ulb.leave_bonus_type_id, ulb.year_belong AS year, ulb.created_at, " +
		"up2.first_name || ' ' || up2.last_name created_by " +
		"FROM user_leave_bonus AS ulb " +
		"JOIN user_profiles AS up ON up.user_id = ulb.user_id " +
		"JOIN user_profiles AS up2 ON up2.user_id = ulb.created_by "

	if params.FullName != "" {
		condition += "WHERE vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) " +
			"LIKE vietnamese_unaccent(LOWER('%" + params.FullName + "%')) "
	}

	if params.LeaveBonusTypeId != 0 {
		condition += "AND ulb.leave_bonus_type_id = " + strconv.Itoa(params.LeaveBonusTypeId)
	}

	if params.Year != 0 {
		condition += " AND ulb.year_belong = " + strconv.Itoa(params.Year)
	}

	if params.IsDeleted {
		condition += " AND ulb.deleted_at IS NOT NULL"
	} else {
		condition += " AND ulb.deleted_at IS NULL"
	}
	condition += " AND ulb.organization_id = " + strconv.Itoa(orgID)

	q1 += condition
	q2 += condition

	_, err := repo.DB.Query(&totalRow, q1)
	if err != nil {
		repo.Logger.Error(err)
		return nil, 0, err
	}

	q2 += " ORDER BY ulb.created_at DESC OFFSET " + strconv.Itoa((params.CurrentPage-1)*params.RowPerPage) +
		" LIMIT " + strconv.Itoa(params.RowPerPage)
	_, err = repo.DB.Query(&records, q2)
	if err != nil {
		repo.Logger.Error(err)
		return nil, 0, err
	}

	return records, totalRow, err
}

func (repo *PgLeaveRepository) SelectLeaveBonusById(Id int) (m.UserLeaveBonus, error) {
	var leaveBonus m.UserLeaveBonus
	err := repo.DB.Model(&leaveBonus).
		Column("leave_bonus_type_id", "reason", "hour", "year_belong").
		Where("id = ?", Id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return leaveBonus, err
}

func (repo *PgLeaveRepository) UpdateLeaveBonus(userId int, params *param.EditLeaveBonusParam) error {
	leaveBonus := m.UserLeaveBonus{
		LeaveBonusTypeID: params.LeaveBonusTypeId,
		YearBelong:       params.YearBelong,
		UpdatedBy:        userId,
		Reason:           params.Reason,
		Hour:             params.Hour,
		HourRemaining:    params.Hour,
	}

	_, err := repo.DB.Model(&leaveBonus).
		Where("id = ?", params.Id).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) UpdateDeleted(id int, isDeleted bool) error {
	q := "UPDATE user_leave_bonus AS ulb SET deleted_at = "
	if isDeleted {
		q += "'" + utils.TimeNowUTC().Format(cf.FormatDate) + "'"
	} else {
		q += "NULL"
	}

	q += " WHERE id = " + strconv.Itoa(id)
	_, err := repo.DB.Query(m.UserLeaveBonus{}, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) SearchUserLeave(orgID int, leaveHistoryParams *param.LeaveHistoryParams) ([]m.UserLeaveRequestExt, int, error) {
	var userLeaveRequestExt []m.UserLeaveRequestExt

	queryObj := repo.DB.Model(&userLeaveRequestExt).
		Relation("UserProfile").
		Relation("UserLeaveRequest", func(q *orm.Query) (*orm.Query, error) {
			return q.Where("DATE(ulr.datetime_leave_from) IN (?)", pg.In(leaveHistoryParams.DateOfWeek)), nil
		}).
		Where("usr.organization_id = ?", orgID).
		Where("user_profile.status != 3")

	if leaveHistoryParams.ID != 0 {
		queryObj.Where("id = ?", leaveHistoryParams.ID)
	}

	if leaveHistoryParams.UserID != 0 {
		queryObj.Where("usr.id = ?", leaveHistoryParams.UserID)
	}

	if leaveHistoryParams.UserName != "" {
		userName := "%" + leaveHistoryParams.UserName + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(user_profile.first_name)) || ' ' || vietnamese_unaccent(LOWER(user_profile.last_name)) LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if leaveHistoryParams.DatetimeLeaveFrom != "" ||
		leaveHistoryParams.DatetimeLeaveTo != "" ||
		leaveHistoryParams.SubtractDayOffTypeID != 0 {
		qstr := `exists (
		                       select  1
		                       from    user_leave_requests ulr
		                       where   ulr.user_id = usr.id`
		if leaveHistoryParams.DatetimeLeaveFrom != "" {
			qstr += ` and     date(ulr.datetime_leave_from) >= DATE('` + leaveHistoryParams.DatetimeLeaveFrom + `')`
		}

		if leaveHistoryParams.DatetimeLeaveTo != "" {
			qstr += ` and     date(ulr.datetime_leave_from) <= DATE('` + leaveHistoryParams.DatetimeLeaveTo + `')`
		}

		if leaveHistoryParams.SubtractDayOffTypeID != 0 {
			qstr += ` and     subtract_day_off_type_id = ` + strconv.Itoa(leaveHistoryParams.SubtractDayOffTypeID)
		}

		qstr += `)`

		queryObj.Where(qstr, leaveHistoryParams.DatetimeLeaveTo)
	}

	queryObj.Offset((leaveHistoryParams.CurrentPage - 1) * leaveHistoryParams.RowPerPage)
	queryObj.Limit(leaveHistoryParams.RowPerPage)
	totalRow, err := queryObj.SelectAndCount()

	if err != nil {
		repo.Logger.Error(err)
	}

	return userLeaveRequestExt, totalRow, err
}

// CountExpireDate : Sum hour remaining of user
func (repo *PgLeaveRepository) CountExpireDate(orgID int) ([]param.ExpireLeaveBonusRecords, error) {
	var expireLeaveBonusRecords []param.ExpireLeaveBonusRecords
	now := time.Now()

	err := repo.DB.Model(&m.UserLeaveBonus{}).
		ColumnExpr("SUM(hour_remaining) as total_hour").
		Column("user_id").
		Where("organization_id = ?", orgID).
		Where("date(ulb.expire_bonus_leave_date) <= DATE(?)", now).
		Where("ulb.hour_remaining is not null").
		Group("user_id").
		Select(&expireLeaveBonusRecords)

	return expireLeaveBonusRecords, err
}

func (repo *PgLeaveRepository) ClearExpireDate(orgID int) error {
	leaveBonus := m.UserLeaveBonus{
		HourRemaining: 0,
	}

	now := time.Now()
	_, err := repo.DB.Model(&leaveBonus).
		Column("hour_remaining").
		Where("organization_id = ?", orgID).
		Where("date(ulb.expire_bonus_leave_date) <= DATE(?)", now).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) SelectValidDateLeave(orgID int, userID int) ([]param.ValidLeaveBonusRecords, error) {
	var userLeaveBonus []m.UserLeaveBonus
	var validLeaveBonusRecords []param.ValidLeaveBonusRecords

	now := time.Now()
	queryObj := repo.DB.Model(&userLeaveBonus)
	queryObj.Column("ulb.id", "ulb.hour", "ulb.hour_remaining").
		Where("user_id = ?", userID).
		Where("organization_id = ?", orgID).
		Where("date(ulb.expire_bonus_leave_date) >= DATE(?)", now).
		Where("ulb.hour_remaining is not null").
		Order("ulb.expire_bonus_leave_date ASC")

	err := queryObj.Select(&validLeaveBonusRecords)
	if err != nil {
		repo.Logger.Error(err)
	}

	return validLeaveBonusRecords, err
}

func (repo *PgLeaveRepository) SelectValidDateOfUser(orgID int, userID int) ([]param.ValidLeaveBonusRecords, error) {
	var userLeaveBonus []m.UserLeaveBonus
	var validLeaveBonusRecords []param.ValidLeaveBonusRecords

	now := time.Now()
	queryObj := repo.DB.Model(&userLeaveBonus)
	queryObj.Column("ulb.id", "ulb.hour", "ulb.hour_remaining").
		Where("user_id = ?", userID).
		Where("organization_id = ?", orgID).
		Where("date(ulb.expire_bonus_leave_date) >= DATE(?)", now).
		Order("ulb.expire_bonus_leave_date ASC")

	err := queryObj.Select(&validLeaveBonusRecords)
	if err != nil {
		repo.Logger.Error(err)
	}

	return validLeaveBonusRecords, err
}

func (repo *PgLeaveRepository) UpdateHourRemainingLeaveByID(id int, hourRemaining float64) error {
	leaveBonus := m.UserLeaveBonus{
		HourRemaining: hourRemaining,
	}

	_, err := repo.DB.Model(&leaveBonus).
		Column("hour_remaining").
		Where("id = ?", id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgLeaveRepository) SelectStartLeaveCurrentYear(orgID int) ([]param.StartDateLeaveRecords, error) {
	var userLeaveBonus []m.UserLeaveBonus
	var startDateLeaveRecords []param.StartDateLeaveRecords

	now := time.Now()
	queryObj := repo.DB.Model(&userLeaveBonus)
	queryObj.Column("ulb.id", "ulb.updated_at", "ulb.leave_bonus_type_id", "ulb.hour").
		Where("organization_id = ?", orgID).
		Where("ulb.year_belong = ?", now.Year()).
		Where("ulb.leave_bonus_type_id = 1 or ulb.leave_bonus_type_id = 2 or ulb.leave_bonus_type_id = 8")

	err := queryObj.Select(&startDateLeaveRecords)
	if err != nil {
		repo.Logger.Error(err)
	}

	return startDateLeaveRecords, err
}

func (repo *PgLeaveRepository) UpdateExpireLeaveDate(id int, hour float64, expireLeaveDate string) error {
	leaveBonus := m.UserLeaveBonus{
		ExpireBonusLeaveDate: expireLeaveDate,
		HourRemaining:        hour,
	}

	_, err := repo.DB.Model(&leaveBonus).
		Column("expire_bonus_leave_date", "hour_remaining").
		Where("id = ?", id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
