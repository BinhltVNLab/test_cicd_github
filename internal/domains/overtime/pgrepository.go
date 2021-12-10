package overtime

import (
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

// PgOvertimeRepository : Struct repository
type PgOvertimeRepository struct {
	cm.AppRepository
}

// NewPgOvertimeRepository : Init repository
func NewPgOvertimeRepository(logger echo.Logger) (repo *PgOvertimeRepository) {
	repo = &PgOvertimeRepository{}
	repo.Init(logger)
	return
}

func (repo *PgOvertimeRepository) InsertOvertimeRequest(
	createOvertimeParams *param.CreateOvertimeParams,
	organizationId int,
	notificationRepo rp.NotificationRepository,
	userRepo rp.UserRepository,
	uniqueUsersId []int,
) (string, string, error) {
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		userOvertimeRequest := m.UserOvertimeRequest{
			UserId:               createOvertimeParams.UserId,
			ProjectId:            createOvertimeParams.ProjectId,
			Status:               createOvertimeParams.Status,
			DatetimeOvertimeFrom: calendar.ParseTime(cf.FormatDateNoSec, createOvertimeParams.DatetimeOvertimeFrom),
			DatetimeOvertimeTo:   calendar.ParseTime(cf.FormatDateNoSec, createOvertimeParams.DatetimeOvertimeTo),
			EmailTitle:           createOvertimeParams.EmailTitle,
			EmailContent:         createOvertimeParams.EmailContent,
			Reason:               createOvertimeParams.Reason,
			OvertimeType:         createOvertimeParams.OvertimeType,
			SendTo:               createOvertimeParams.SendTo,
			SendCc:               createOvertimeParams.SendCc,
			WorkAtNoon:           createOvertimeParams.WorkAtNoon,
		}

		transErr = tx.Insert(&userOvertimeRequest)
		if transErr != nil {
			return transErr
		}

		notificationParams := new(param.InsertNotificationParam)
		notificationParams.Content = "has just created a overtime request"
		notificationParams.RedirectUrl = "/request/manage-overtime?id=" + strconv.Itoa(userOvertimeRequest.ID)

		for _, userId := range uniqueUsersId {
			if userId == createOvertimeParams.UserId {
				continue
			}
			notificationParams.Receiver = userId
			transErr = notificationRepo.InsertNotificationWithTx(tx, organizationId, createOvertimeParams.UserId, notificationParams)
			if transErr != nil {
				return transErr
			}
		}

		body = notificationParams.Content
		link = notificationParams.RedirectUrl

		return transErr
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}

func (repo *PgOvertimeRepository) UpdateStatusOvertimeRequest(
	updateRequestStatusParams *param.UpdateRequestStatusParams,
	notificationRepo rp.NotificationRepository,
	leaveRepo rp.LeaveRepository,
	userRepo rp.UserRepository,
	organizationId int,
	userId int,
	hour float64,
) (string, string, error) {
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		_, transErr = tx.Model(&m.UserOvertimeRequest{Status: updateRequestStatusParams.Status}).
			Column("status", "updated_at").
			Where("id = ?", updateRequestStatusParams.RequestID).
			Update()

		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		userOvertimeRequest, transErr := repo.SelectOvertimeRequestById(updateRequestStatusParams.RequestID)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		notificationParams := new(param.InsertNotificationParam)
		notificationParams.Receiver = userOvertimeRequest.UserId
		if updateRequestStatusParams.Status == cf.AcceptRequestStatus {
			notificationParams.Content = "has just accepted a overtime request"
		} else {
			notificationParams.Content = "has just denied a overtime request"
		}
		notificationParams.RedirectUrl = "/request/manage-overtime?id=" + strconv.Itoa(updateRequestStatusParams.RequestID)
		transErr = notificationRepo.InsertNotificationWithTx(tx, organizationId, userId, notificationParams)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		now := time.Now()
		if updateRequestStatusParams.Status == cf.AcceptRequestStatus && userOvertimeRequest.OvertimeType == cf.DayOffTypeOvertime {
			leaveBonusParams := param.LeaveBonus{
				OrgID:            organizationId,
				UserID:           userOvertimeRequest.UserId,
				LeaveBonusTypeID: cf.OvertimeLeave,
				CreatedBy:        userId,
				UpdatedBy:        userId,
				YearBelong:       userOvertimeRequest.DatetimeOvertimeFrom.Year(),
				Reason:           userOvertimeRequest.Reason,
				Hour:             hour,
				HourRemaining:    hour,
				ExpireBonusLeaveDate: now.AddDate(0, 3, 0).Format("2006-01-02 15:04:05"),
			}
			transErr = leaveRepo.InsertLeaveBonusOvertimeWithTx(tx, &leaveBonusParams)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		body = notificationParams.Content
		link = notificationParams.RedirectUrl

		return transErr
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}

func (repo *PgOvertimeRepository) SelectOvertimeRequests(
	organizationId int,
	getOvertimeRequestsParams *param.GetOvertimeRequestsParams,
) ([]param.OvertimeRequestsRecords, int, error) {
	var records []param.OvertimeRequestsRecords
	q := repo.DB.Model(&m.UserOvertimeRequest{})
	q.Column("uotr.id", "uotr.status", "uotr.datetime_overtime_from", "uotr.datetime_overtime_to",
		"uotr.overtime_type", "uotr.work_at_noon", "up.employee_id").
		ColumnExpr("EXTRACT(HOUR FROM datetime_overtime_from) AS hour_from").
		ColumnExpr("EXTRACT(MINUTE FROM datetime_overtime_from) AS minute_from").
		ColumnExpr("EXTRACT(HOUR FROM datetime_overtime_to) AS hour_to").
		ColumnExpr("EXTRACT(MINUTE FROM datetime_overtime_to) AS minute_to").
		ColumnExpr("p.name AS project_name").
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		ColumnExpr("b.name AS branch").
		Join("JOIN user_profiles AS up ON up.user_id = uotr.user_id").
		Join("JOIN users AS u ON u.id = uotr.user_id").
		Join("LEFT JOIN branches AS b ON b.id = up.branch").
		Join("JOIN projects AS p ON p.id = uotr.project_id").
		Where("u.organization_id = ?", organizationId)

	if getOvertimeRequestsParams.Id != 0 {
		q.Where("uotr.id = ?", getOvertimeRequestsParams.Id)
	}

	if len(getOvertimeRequestsParams.UsersId) == 1 {
		q.Where("uotr.user_id = ?", getOvertimeRequestsParams.UsersId[0])
	} else if len(getOvertimeRequestsParams.UsersId) > 1 {
		q.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, userId := range getOvertimeRequestsParams.UsersId {
				q = q.WhereOr("uotr.user_id = ?", userId)
			}
			return q, nil
		})
	}

	if getOvertimeRequestsParams.OvertimeType != 0 {
		q.Where("uotr.overtime_type = ?", getOvertimeRequestsParams.OvertimeType)
	}

	if getOvertimeRequestsParams.Status != 0 {
		q.Where("uotr.status = ?", getOvertimeRequestsParams.Status)
	}

	if getOvertimeRequestsParams.ProjectId != 0 {
		q.Where("p.id = ?", getOvertimeRequestsParams.ProjectId)
	}

	if getOvertimeRequestsParams.Branch != 0 {
		q.Where("b.id = ?", getOvertimeRequestsParams.Branch)
	}

	if getOvertimeRequestsParams.DateFrom != "" {
		q.Where("DATE(uotr.datetime_overtime_from) >= to_date(?,'YYYY-MM-DD')", getOvertimeRequestsParams.DateFrom)
	}

	if getOvertimeRequestsParams.DateTo != "" {
		q.Where("DATE(uotr.datetime_overtime_to) <= to_date(?,'YYYY-MM-DD')", getOvertimeRequestsParams.DateTo)
	}

	q.OrderExpr("uotr.datetime_overtime_from DESC").
		Offset((getOvertimeRequestsParams.CurrentPage - 1) * getOvertimeRequestsParams.RowPerPage).
		Limit(getOvertimeRequestsParams.RowPerPage)

	totalRow, err := q.SelectAndCount(&records)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgOvertimeRepository) SelectOvertimeRequestById(id int) (m.UserOvertimeRequest, error) {
	var userOvertimeRequest m.UserOvertimeRequest
	err := repo.DB.Model(&userOvertimeRequest).
		Column("user_id", "project_id", "datetime_overtime_from", "datetime_overtime_to", "status",
			"email_title", "email_content", "reason", "overtime_type", "work_at_noon", "send_to", "send_cc").
		Where("id = ?", id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return userOvertimeRequest, err
}

func (repo *PgOvertimeRepository) UpdateOvertimeRequest(params *param.UpdateOvertimeRequestParams) error {
	userOvertimeRequest := m.UserOvertimeRequest{
		ProjectId:            params.ProjectId,
		DatetimeOvertimeFrom: calendar.ParseTime(cf.FormatDateNoSec, params.DatetimeOvertimeFrom),
		DatetimeOvertimeTo:   calendar.ParseTime(cf.FormatDateNoSec, params.DatetimeOvertimeTo),
		EmailTitle:           params.EmailTitle,
		EmailContent:         params.EmailContent,
		Reason:               params.Reason,
		OvertimeType:         params.OvertimeType,
		SendTo:               params.SendTo,
		SendCc:               params.SendCc,
	}

	_, err := repo.DB.Model(&userOvertimeRequest).
		Column("project_id", "datetime_overtime_from", "datetime_overtime_to", "email_title",
			"email_content", "reason", "overtime_type", "send_to", "send_cc", "updated_at").
		Where("id = ?", params.Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgOvertimeRepository) InsertOvertimeWeight(
	organizationId int,
	params *param.CreateOvertimeWeightParams,
	settingStep int,
	orgRepo rp.OrgRepository,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		overtimeWeight := m.OvertimeWeight{
			OrganizationId:  organizationId,
			NormalDayWeight: params.NormalDayWeight,
			WeekendWeight:   params.WeekendWeight,
			HolidayWeight:   params.HolidayWeight,
		}

		transErr = tx.Insert(&overtimeWeight)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		if settingStep == 0 {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.ORGANIZATIONEMAILSETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		} else if settingStep < cf.OVERTIMESETTING {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.OVERTIMESETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		return transErr
	})

	return err
}

func (repo *PgOvertimeRepository) UpdateOvertimeWeight(params *param.EditOvertimeWeightParams) error {
	overtimeWeight := m.OvertimeWeight{
		NormalDayWeight: params.NormalDayWeight,
		WeekendWeight:   params.WeekendWeight,
		HolidayWeight:   params.HolidayWeight,
	}

	_, err := repo.DB.Model(&overtimeWeight).
		Column("normal_day_weight", "weekend_weight", "holiday_weight", "updated_at").
		Where("id = ?", params.Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgOvertimeRepository) CountOvertimeWeightByField(field string, value int) (int, error) {
	var overtimeWeight m.OvertimeWeight
	count, err := repo.DB.Model(&overtimeWeight).Where(field+" = ?", value).Count()
	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgOvertimeRepository) SelectOvertimeWeightByOrganizationId(organizationId int) (m.OvertimeWeight, error) {
	var overtimeWeight m.OvertimeWeight
	err := repo.DB.Model(&overtimeWeight).
		Column("id", "normal_day_weight", "weekend_weight", "holiday_weight").
		Where("organization_id = ?", organizationId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return overtimeWeight, err
}
