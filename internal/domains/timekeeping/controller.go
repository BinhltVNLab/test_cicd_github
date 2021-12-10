package timekeeping

import (
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"

	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// TkController : Timekeeping Controller
type TkController struct {
	cm.BaseController

	TimekeepingRepo rp.TimekeepingRepository
	UserRepo        rp.UserRepository
	BranchRepo      rp.BranchRepository
}

// NewTimekeepingController : Init Timekeeping Controller
func NewTimekeepingController(logger echo.Logger, timekeepingRepo rp.TimekeepingRepository, userRepo rp.UserRepository, branchRepo rp.BranchRepository) (ctr *TkController) {
	ctr = &TkController{cm.BaseController{}, timekeepingRepo, userRepo, branchRepo}
	ctr.Init(logger)
	return
}

// CheckIn : Check in time
func (ctr *TkController) CheckIn(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	orgID, userID := userProfile.OrganizationID, userProfile.UserProfile.UserID

	timekeeping, err := ctr.TimekeepingRepo.GetLastTimekeepingToday(orgID, userID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if !timekeeping.CheckInTime.IsZero() && timekeeping.CheckOutTime.IsZero() {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You have not check out for previous time",
		})
	}

	err = ctr.TimekeepingRepo.InsertCheckInTime(orgID, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Check in successfully.",
	})
}

// CheckOut : Check out time
func (ctr *TkController) CheckOut(c echo.Context) error {

	userProfile := c.Get("user_profile").(m.User)

	orgID, userID := userProfile.OrganizationID, userProfile.UserProfile.UserID

	timekeeping, err := ctr.TimekeepingRepo.GetLastTimekeepingToday(orgID, userID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if (timekeeping.CheckInTime.IsZero() && timekeeping.CheckOutTime.IsZero()) ||
		(!timekeeping.CheckInTime.IsZero() && !timekeeping.CheckOutTime.IsZero()) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You have check-in before check-out",
		})
	}

	err = ctr.TimekeepingRepo.InsertCheckOutTime(timekeeping.ID)

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Check out successfully.",
	})
}

// GetTimekeeping : Get timekeeping
func (ctr *TkController) GetTimekeeping(c echo.Context) error {

	userProfile := c.Get("user_profile").(m.User)

	orgID, userID := userProfile.OrganizationID, userProfile.ID

	timekeeping, err := ctr.TimekeepingRepo.GetLastTimekeepingToday(orgID, userID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	utcLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	currentTime := utils.TimeNowUTC()

	checkinTimeVN := ""
	if !timekeeping.CheckInTime.IsZero() {
		checkinTimeVN = timekeeping.CheckInTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
	}

	checkoutTimeVN := ""
	if !timekeeping.CheckOutTime.IsZero() {
		checkoutTimeVN = timekeeping.CheckOutTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
	}

	dataResponse := map[string]interface{}{
		"check_in_time":  checkinTimeVN,
		"check_out_time": checkoutTimeVN,
		"time_server":    currentTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping2),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get timekeeping successfully.",
		Data:    dataResponse,
	})
}

// GetAllTimekeepingUser : Get all timekeeping of user
func (ctr *TkController) GetAllTimekeepingUser(c echo.Context) error {
	seachTimekeepingUserParams := new(param.SeachTimekeepingUserParams)
	seachTimekeepingUserParams.RowPerPage = 10

	if err := c.Bind(seachTimekeepingUserParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	orgID, userID := userProfile.OrganizationID, userProfile.UserProfile.UserID

	timekeepings, totalRow, err := ctr.TimekeepingRepo.GetAllTimekeepingUser(orgID, userID, seachTimekeepingUserParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if seachTimekeepingUserParams.RowPerPage == 0 {
		seachTimekeepingUserParams.CurrentPage = 1
		seachTimekeepingUserParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": seachTimekeepingUserParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": seachTimekeepingUserParams.RowPerPage,
	}

	utcLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	timekeepingsResponse := []map[string]interface{}{}

	for i := 0; i < len(timekeepings); i++ {
		checkinTimeVN := ""
		if !timekeepings[i].CheckInTime.IsZero() {
			checkinTimeVN = timekeepings[i].CheckInTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
		}

		checkoutTimeVN := ""
		if !timekeepings[i].CheckOutTime.IsZero() {
			checkoutTimeVN = timekeepings[i].CheckOutTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
		}

		timekeeping := map[string]interface{}{
			"check_in_time":  checkinTimeVN,
			"check_out_time": checkoutTimeVN,
		}

		timekeepingsResponse = append(timekeepingsResponse, timekeeping)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination":   pagination,
			"timekeepings": timekeepingsResponse,
		},
	})
}

// GetAllTimekeeping : Get timekeeping all user
func (ctr *TkController) GetAllTimekeeping(c echo.Context) error {
	seachAllTimekeepingParams := new(param.SeachAllTimekeepingParams)
	seachAllTimekeepingParams.RowPerPage = 10

	if err := c.Bind(seachAllTimekeepingParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	orgID := userProfile.OrganizationID

	timekeepings, totalRow, err := ctr.TimekeepingRepo.GetAllTimekeeping(orgID, seachAllTimekeepingParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get timekeeping list failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if seachAllTimekeepingParams.RowPerPage == 0 {
		seachAllTimekeepingParams.CurrentPage = 1
		seachAllTimekeepingParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": seachAllTimekeepingParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": seachAllTimekeepingParams.RowPerPage,
	}

	timekeepingsResponse := []map[string]interface{}{}
	utcLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	for _, record := range branchRecords {
		branches[record.Id] = record.Name
	}

	for i := 0; i < len(timekeepings); i++ {
		checkinTimeVN := ""
		if !timekeepings[i].CheckInTime.IsZero() {
			checkinTimeVN = timekeepings[i].CheckInTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
		}

		checkoutTimeVN := ""
		if !timekeepings[i].CheckOutTime.IsZero() {
			checkoutTimeVN = timekeepings[i].CheckOutTime.In(utcLocation).Format(cf.FormatDisplayTimekeeping)
		}

		timekeeping := map[string]interface{}{
			"user_name":      timekeepings[i].UserName,
			"email":          timekeepings[i].Email,
			"branch":         branches[timekeepings[i].Branch],
			"check_in_time":  checkinTimeVN,
			"check_out_time": checkoutTimeVN,
		}

		timekeepingsResponse = append(timekeepingsResponse, timekeeping)
	}

	users, err := ctr.UserRepo.GetAllUserNameByOrgID(orgID)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get user list failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userList := make(map[int]string)
	for i := 0; i < len(users); i++ {
		userList[users[i].UserID] = users[i].FullName
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"branch_select_box": branches,
			"users_select_box":  userList,
			"pagination":        pagination,
			"timekeepings":      timekeepingsResponse,
		},
	})
}

// ExportExcel : Export to CSV
func (ctr *TkController) ExportExcel(c echo.Context) error {
	exportExcelParams := &param.TkExportExcelParams{
		DateFrom: c.FormValue("date_from"),
		DateTo:   c.FormValue("date_to"),
	}

	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.TimekeepingRepo.SelectTimekeepingsByDate(userProfile.OrganizationID, exportExcelParams)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Empty records.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "A", 30)
	_ = f.SetColWidth("Sheet1", "B", "D", 15)

	titleStyle, _ := f.NewStyle(`{
		"font":{"bold":true, "size":16},
		"alignment":{"horizontal":"center", "vertical":"center"}
	}`)

	contentStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center", "vertical":"center"}}`)

	categories := map[string]string{"A1": "Full name", "B1": "Date", "C1": "Check in", "D1": "Check out"}
	for k, v := range categories {
		_ = f.SetCellValue("Sheet1", k, v)
	}
	_ = f.SetCellStyle("Sheet1", "A1", "D1", titleStyle)
	_ = f.SetColStyle("Sheet1", "B", contentStyle)
	_ = f.SetColStyle("Sheet1", "C", contentStyle)
	_ = f.SetColStyle("Sheet1", "D", contentStyle)

	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	for i, record := range records {
		firstTime, lastTime := "", ""
		if !record.CheckInTime.IsZero() {
			firstTime = utils.ConvertTwoChar(record.CheckInTime.In(location).Hour()) + ":" + utils.ConvertTwoChar(record.CheckInTime.In(location).Minute())
		}

		if !record.CheckOutTime.IsZero() {
			lastTime = utils.ConvertTwoChar(record.CheckOutTime.In(location).Hour()) + ":" + utils.ConvertTwoChar(record.CheckOutTime.In(location).Minute())
		}

		pos := i + 2
		values := map[string]interface{}{
			"A" + strconv.Itoa(pos): record.FullName,
			"B" + strconv.Itoa(pos): record.Date.In(location).Format(cf.FormatDateDisplay),
			"C" + strconv.Itoa(pos): firstTime,
			"D" + strconv.Itoa(pos): lastTime,
		}

		for k, v := range values {
			_ = f.SetCellValue("Sheet1", k, v)
		}
	}

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}
