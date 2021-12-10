package timekeeping

import (
	"github.com/labstack/echo/v4"
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// PgTimekeepingRepository : Timekeeping Repository
type PgTimekeepingRepository struct {
	cm.AppRepository
}

// NewPgTimekeepingRepository : Init PgTimekeepingRepository
func NewPgTimekeepingRepository(logger echo.Logger) (repo *PgTimekeepingRepository) {
	repo = &PgTimekeepingRepository{}
	repo.Init(logger)
	return
}

// InsertCheckInTime : Insert check in time record to database
func (repo *PgTimekeepingRepository) InsertCheckInTime(orgID int, userID int) error {
	checkInTime := m.UserTimekeeping{
		OrganizationID: orgID,
		UserID:         userID,
		CheckInTime:    utils.TimeNowUTC(),
	}

	err := repo.DB.Insert(&checkInTime)

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// InsertCheckOutTime : Insert check out time record to database
func (repo *PgTimekeepingRepository) InsertCheckOutTime(ID int) error {
	_, err := repo.DB.Model(&m.UserTimekeeping{CheckOutTime: utils.TimeNowUTC()}).
		Column("check_out_time", "updated_at").
		Where("id = ?", ID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// GetTimekeepingWithMaxCheckInTime : Get Timekeeping with max check_in_time
// Param                     : OrgID, userID
// Return                    : m.UserTimekeeping, error
func (repo *PgTimekeepingRepository) GetTimekeepingWithMaxCheckInTime(orgID int, userID int) (m.UserTimekeeping, error) {
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	timekeeping := m.UserTimekeeping{}
	err := repo.DB.Model(&timekeeping).
		Column("id, user_id, organization_id, check_in_time, check_out_time").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("date(check_in_time at time zone 'utc' at time zone 'Asia/Ho_Chi_Minh') = date(?)", utils.TimeNowUTC().In(location)).
		Order("check_in_time DESC").
		Limit(1).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return timekeeping, err
}

// GetLastTimekeepingToday : Get Timekeeping with max check_in_time & check_out_time
// Param                     : OrgID, userID
// Return                    : param.AllUserTimekeepingResponse, error
func (repo *PgTimekeepingRepository) GetLastTimekeepingToday(orgID int, userID int) (m.UserTimekeeping, error) {
	timekeeping := m.UserTimekeeping{}

	err := repo.DB.Model(&timekeeping).
		Column("id").
		Column("organization_id").
		Column("user_id").
		Column("check_in_time").
		Column("check_out_time").
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("date(check_in_time) = date(?)", utils.TimeNowUTC()).
		Order("check_in_time DESC").
		Limit(1).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return timekeeping, err
}

// GetAllTimekeepingUser : Get all Timekeeping of user
func (repo *PgTimekeepingRepository) GetAllTimekeepingUser(
	orgID int,
	userID int,
	seachTimekeepingUserParams *param.SeachTimekeepingUserParams,
) ([]param.UserTimekeepingResponse, int, error) {
	timekeepings := []m.UserTimekeeping{}
	records := []param.UserTimekeepingResponse{}

	queryObj := repo.DB.Model(&timekeepings)
	queryObj.Column("id")
	queryObj.Column("user_id")
	queryObj.Column("organization_id")
	queryObj.Column("check_in_time")
	queryObj.Column("check_out_time")
	queryObj.Where("organization_id = ?", orgID)
	queryObj.Where("user_id = ?", userID)

	if seachTimekeepingUserParams.DateFrom != "" {
		queryObj.Where("date(check_in_time) >= to_date(?,'YYYY-MM-DD')", seachTimekeepingUserParams.DateFrom)
	}

	if seachTimekeepingUserParams.DateTo != "" {
		queryObj.Where("date(check_in_time) <= to_date(?,'YYYY-MM-DD')", seachTimekeepingUserParams.DateTo)
	}

	queryObj.Offset((seachTimekeepingUserParams.CurrentPage - 1) * seachTimekeepingUserParams.RowPerPage)
	queryObj.Order("check_in_time DESC")
	queryObj.Limit(seachTimekeepingUserParams.RowPerPage)

	totalRow, err := queryObj.SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

// GetAllTimekeeping : Get all Timekeeping
func (repo *PgTimekeepingRepository) GetAllTimekeeping(
	orgID int,
	seachAllTimekeepingParams *param.SeachAllTimekeepingParams,
) ([]param.UserTimekeepingResponse, int, error) {
	timekeepings := []m.UserTimekeeping{}
	records := []param.UserTimekeepingResponse{}

	queryObj := repo.DB.Model(&timekeepings)
	queryObj.Column("utk.organization_id")
	queryObj.Column("utk.user_id")
	queryObj.ColumnExpr("min(usr.first_name || ' ' || usr.last_name) as user_name")
	queryObj.ColumnExpr("uss.email as email")
	queryObj.ColumnExpr("date(utk.check_in_time) as date_timekeeping")
	queryObj.ColumnExpr("min(utk.check_in_time) as check_in_time")
	queryObj.ColumnExpr("max(utk.check_out_time) as check_out_time")
	queryObj.ColumnExpr("min(usr.branch) as branch")
	queryObj.Join("join user_profiles as usr on usr.user_id = utk.user_id")
	queryObj.Join("join users as uss on uss.id = utk.user_id")
	queryObj.Where("utk.organization_id = ?", orgID)
	queryObj.Where("usr.status != 3")

	if seachAllTimekeepingParams.UserName != "" {
		userName := "%" + seachAllTimekeepingParams.UserName + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(usr.first_name)) || ' ' || vietnamese_unaccent(LOWER(usr.last_name)) LIKE vietnamese_unaccent(LOWER(?0))",
		userName)
	}

	if seachAllTimekeepingParams.BranchID != 0 {
		queryObj.Where("usr.branch = ?", seachAllTimekeepingParams.BranchID)
	}

	if seachAllTimekeepingParams.FromDate != "" {
		queryObj.Where("date(utk.check_in_time) >= to_date(?,'YYYY-MM-DD')", seachAllTimekeepingParams.FromDate)
	}

	if seachAllTimekeepingParams.ToDate != "" {
		queryObj.Where("date(utk.check_in_time) <= to_date(?,'YYYY-MM-DD')", seachAllTimekeepingParams.ToDate)
	}

	queryObj.Group("utk.organization_id", "utk.user_id", "date_timekeeping", "uss.email")
	queryObj.Offset((seachAllTimekeepingParams.CurrentPage - 1) * seachAllTimekeepingParams.RowPerPage)
	queryObj.Limit(seachAllTimekeepingParams.RowPerPage)
	queryObj.Order("date_timekeeping DESC")

	totalRow, err := queryObj.SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgTimekeepingRepository) SelectTimekeepingsByDate(organizationId int, exportExcelParams *param.TkExportExcelParams) ([]param.TkExportExcelRecords, error) {
	var records []param.TkExportExcelRecords
	q := repo.DB.Model(&m.UserTimekeeping{})
	q.ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		ColumnExpr("DATE(utk.check_in_time) AS date").
		ColumnExpr("MIN(utk.check_in_time) AS check_in_time").
		ColumnExpr("MAX(utk.check_out_time) AS check_out_time").
		Join("JOIN user_profiles AS up ON up.user_id = utk.user_id").
		Where("utk.organization_id = ?", organizationId)

	if exportExcelParams.DateFrom != "" {
		q.Where("DATE(utk.check_in_time) >= to_date(?,'YYYY-MM-DD')", exportExcelParams.DateFrom)
	}

	if exportExcelParams.DateTo != "" {
		q.Where("DATE(utk.check_out_time) <= to_date(?,'YYYY-MM-DD')", exportExcelParams.DateTo)
	}

	q.GroupExpr("full_name, date")
	q.OrderExpr("full_name, date ASC")

	err := q.Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}
