package overtime

import (
	"net/http"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type Controller struct {
	cm.BaseController
	email.SMTPGoMail
	afb.FirebaseCloudMessage

	OvertimeRepo     rp.OvertimeRepository
	OrgRepo          rp.OrgRepository
	UserRepo         rp.UserRepository
	ProjectRepo      rp.ProjectRepository
	BranchRepo       rp.BranchRepository
	leaveRepo        rp.LeaveRepository
	HolidayRepo      rp.HolidayRepository
	NotificationRepo rp.NotificationRepository
	UserProjectRepo  rp.UserProjectRepository
	FcmTokenRepo     rp.FcmTokenRepository
}

// NewOvertimeController : Init Overtime Controller
func NewOvertimeController(
	logger echo.Logger,
	overtimeRepo rp.OvertimeRepository,
	orgRepo rp.OrgRepository,
	userRepo rp.UserRepository,
	projectRepo rp.ProjectRepository,
	branchRepo rp.BranchRepository,
	leaveRepo rp.LeaveRepository,
	holidayRepo rp.HolidayRepository,
	notificationRepo rp.NotificationRepository,
	userProjectRepo rp.UserProjectRepository,
	fcmTokenRepo rp.FcmTokenRepository,
) (ctr *Controller) {
	ctr = &Controller{
		cm.BaseController{}, email.SMTPGoMail{}, afb.FirebaseCloudMessage{},
		overtimeRepo, orgRepo, userRepo, projectRepo,
		branchRepo, leaveRepo, holidayRepo, notificationRepo,
		userProjectRepo, fcmTokenRepo,
	}
	ctr.Init(logger)
	ctr.InitFcm()

	return
}

func (ctr *Controller) CreateOvertimeRequest(c echo.Context) error {
	createOvertimeParams := new([]param.CreateOvertimeParams)
	if err := c.Bind(createOvertimeParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	org, err := ctr.OrgRepo.SelectEmailAndPassword(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}
	ctr.InitSmtp(org.Email, org.EmailPassword)

	usersIdGmAndManager, err := ctr.UserRepo.SelectIdsOfGMAndManager(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var uniqueUsersId []int
	for _, overtimeRequest := range *createOvertimeParams {
		overtimeRequest.Status = cf.PendingRequestStatus
		_, err := valid.ValidateStruct(overtimeRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
			})
		}

		if overtimeRequest.UserId != userProfile.UserProfile.UserID {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You not have permission to create overtime request",
			})
		}

		if len(overtimeRequest.UsersIdNotification) > 0 || len(usersIdGmAndManager) > 0 {
			uniqueUsersId = utils.AppendUniqueSlice(overtimeRequest.UsersIdNotification, usersIdGmAndManager)
		}

		body, link, err := ctr.OvertimeRepo.InsertOvertimeRequest(
			&overtimeRequest,
			userProfile.OrganizationID,
			ctr.NotificationRepo,
			ctr.UserRepo,
			uniqueUsersId,
		)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if org.Email != "" && org.EmailPassword != "" {
			if overtimeRequest.EmailTitle != "" &&
				overtimeRequest.EmailContent != "" &&
				len(overtimeRequest.SendTo) > 0 {
				sampleData := new(param.SampleData)
				sampleData.SendTo = overtimeRequest.SendTo
				sampleData.SendCc = overtimeRequest.SendCc
				sampleData.Content = overtimeRequest.EmailContent
				if err := ctr.SendMail(
					overtimeRequest.EmailTitle,
					sampleData,
					cf.OvertimeRequestTemplate,
				); err != nil {
					ctr.Logger.Error(err)
					return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: "System error",
					})
				}
			}
		}

		registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(uniqueUsersId, overtimeRequest.UserId)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if len(registrationTokens) > 0 {
			body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
			for _, token := range registrationTokens {
				err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
				if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
					"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
					_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
				}
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create overtime request successful",
	})
}

func (ctr *Controller) UpdateOvertimeRequestStatus(c echo.Context) error {
	updateRequestStatusParams := new(param.UpdateRequestStatusParams)
	if err := c.Bind(updateRequestStatusParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(updateRequestStatusParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if updateRequestStatusParams.Status != cf.AcceptRequestStatus && updateRequestStatusParams.Status != cf.DenyRequestStatus {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Status error",
		})
	}

	userOvertimeRequestBefore, err := ctr.OvertimeRepo.SelectOvertimeRequestById(updateRequestStatusParams.RequestID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "No result",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if userOvertimeRequestBefore.Status == updateRequestStatusParams.Status {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The state of before and after is the same",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	year := userOvertimeRequestBefore.DatetimeOvertimeFrom.Year()
	holidayDates := ctr.getHolidayCurrentYear(userProfile.OrganizationID, year)
	cld := calendar.NewCalendar(holidayDates)
	hour, _, _ := ctr.calculateActualHourOvertime(
		cld,
		userProfile.OrganizationID,
		userOvertimeRequestBefore.DatetimeOvertimeFrom,
		userOvertimeRequestBefore.DatetimeOvertimeTo,
		userOvertimeRequestBefore.WorkAtNoon,
	)
	if hour < 0 {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	body, link, err := ctr.OvertimeRepo.UpdateStatusOvertimeRequest(
		updateRequestStatusParams,
		ctr.NotificationRepo,
		ctr.leaveRepo,
		ctr.UserRepo,
		userProfile.OrganizationID,
		userProfile.UserProfile.UserID,
		hour,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectFcmTokenByUserId(userOvertimeRequestBefore.UserId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if len(registrationTokens) > 0 {
		body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update status request successful",
	})
}

func (ctr *Controller) GetOvertimeRequests(c echo.Context) error {
	getOvertimeRequestsParams := new(param.GetOvertimeRequestsParams)
	if err := c.Bind(getOvertimeRequestsParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID && userProfile.RoleID != cf.ManagerRoleID && len(getOvertimeRequestsParams.UsersId) == 0 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	otRecords, totalRow, err := ctr.OvertimeRepo.SelectOvertimeRequests(userProfile.OrganizationID, getOvertimeRequestsParams)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": getOvertimeRequestsParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getOvertimeRequestsParams.RowPerPage,
	}

	var otResponses []map[string]interface{}
	for _, record := range otRecords {
		holidayDates := ctr.getHolidayCurrentYear(userProfile.OrganizationID, record.DatetimeOvertimeFrom.Year())
		cld := calendar.NewCalendar(holidayDates)
		_, hour, _ := ctr.calculateActualHourOvertime(cld, userProfile.OrganizationID, record.DatetimeOvertimeFrom, record.DatetimeOvertimeTo, record.WorkAtNoon)
		res := map[string]interface{}{
			"id":            record.Id,
			"full_name":     record.FullName,
			"branch":        record.Branch,
			"project_name":  record.ProjectName,
			"status":        utils.GetNameStatusRegistRequests(record.Status),
			"overtime_type": cf.MapOvertimeType[record.OvertimeType],
			"date_overtime": record.DatetimeOvertimeFrom.Format(cf.FormatDateDisplay),
			"time_overtime": utils.ConvertTwoChar(record.HourFrom) + ":" + utils.ConvertTwoChar(record.MinuteFrom) +
				"-" + utils.ConvertTwoChar(record.HourTo) + ":" + utils.ConvertTwoChar(record.MinuteTo),
			"week_day":     record.DatetimeOvertimeFrom.Weekday().String(),
			"working_time": hour,
		}

		otResponses = append(otResponses, res)
	}

	projectManagers, err := ctr.ProjectRepo.SelectProjectManagers(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userList := make(map[int]string)
	if userProfile.RoleID == cf.GeneralManagerRoleID || userProfile.RoleID == cf.ManagerRoleID {
		users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}

		if len(users) > 0 {
			for i := 0; i < len(users); i++ {
				userList[users[i].UserID] = users[i].FullName
			}
		}
	} else if utils.FindIntInSlice(projectManagers, userProfile.UserProfile.UserID) {
		users, err := ctr.UserProjectRepo.SelectUserIdsManagedByManager(userProfile.OrganizationID, userProfile.UserProfile.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}

		if len(users) > 0 {
			for i := 0; i < len(users); i++ {
				userList[users[i].UserId] = users[i].FullName
			}
		}
	}

	projectListParams := param.ProjectListParams{
		CurrentPage: 0,
		Keyword: "",
		RowPerPage: 0,
	}

	projects, _, err := ctr.ProjectRepo.GetProjectsOfUser(userProfile.UserProfile.UserID, userProfile.RoleID, userProfile.OrganizationID, &projectListParams)
	if err != nil && err.Error() == pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	projectList := make(map[int]string)
	if len(projects) > 0 {
		for i := 0; i < len(projects); i++ {
			projectList[projects[i].Id] = projects[i].Name
		}
	}

	branches, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branchList := make(map[int]string)
	for _, record := range branches {
		branchList[record.Id] = record.Name
	}

	responseData := map[string]interface{}{
		"pagination":            pagination,
		"ot_requests":           otResponses,
		"users":                 userList,
		"projects":              projectList,
		"branches":              branchList,
		"overtime_types":        cf.MapOvertimeType,
		"status_overtime_types": cf.MapStatusOvertimeType,
		"project_managers":      projectManagers,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get overtime requests successful",
		Data:    responseData,
	})
}

func (ctr *Controller) GetOvertimeRequestById(c echo.Context) error {
	getOvertimeRequestParam := new(param.GetOvertimeRequestParam)
	if err := c.Bind(getOvertimeRequestParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(getOvertimeRequestParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userOvertimeRequest, err := ctr.OvertimeRepo.SelectOvertimeRequestById(getOvertimeRequestParam.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "No result",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	if userOvertimeRequest.UserId != userProfile.UserProfile.UserID && userProfile.RoleID != cf.GeneralManagerRoleID {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to view overtime",
		})
	}

	dataResponse := map[string]interface{}{
		"user_id":                userOvertimeRequest.UserId,
		"project_id":             userOvertimeRequest.ProjectId,
		"datetime_overtime_from": userOvertimeRequest.DatetimeOvertimeFrom.Format(cf.FormatDateNoSec),
		"datetime_overtime_to":   userOvertimeRequest.DatetimeOvertimeTo.Format(cf.FormatDateNoSec),
		"email_title":            userOvertimeRequest.EmailTitle,
		"email_content":          userOvertimeRequest.EmailContent,
		"reason":                 userOvertimeRequest.Reason,
		"overtime_type":          userOvertimeRequest.OvertimeType,
		"send_to":                userOvertimeRequest.SendTo,
		"send_cc":                userOvertimeRequest.SendCc,
		"work_at_noon":           userOvertimeRequest.WorkAtNoon,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get overtime request successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) EditOvertimeRequest(c echo.Context) error {
	updateOvertimeRequestParams := new(param.UpdateOvertimeRequestParams)
	if err := c.Bind(updateOvertimeRequestParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(updateOvertimeRequestParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userOvertimeRequest, err := ctr.OvertimeRepo.SelectOvertimeRequestById(updateOvertimeRequestParams.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "No result",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if userOvertimeRequest.Status == cf.AcceptRequestStatus || userOvertimeRequest.Status == cf.DenyRequestStatus {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Can't change overtime request",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if updateOvertimeRequestParams.UserId != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to update overtime request",
		})
	}

	err = ctr.OvertimeRepo.UpdateOvertimeRequest(updateOvertimeRequestParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update overtime request successful",
	})
}

func (ctr *Controller) GetEmailsGMAndPM(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.UserRepo.SelectEmailOfGMAndPM(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	emails := make(map[string]string)
	users := make(map[string]string)
	if len(records) > 0 {
		for _, record := range records {
			emails[strconv.Itoa(record.UserId)] = record.Email
			users[strconv.Itoa(record.UserId)] = record.FullName
		}
	}

	dataResponse := map[string]interface{}{
		"emails": emails,
		"users":  users,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get email successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) ExportExcel(c echo.Context) error {
	params := new(param.GetOvertimeRequestsParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	otRecords, _, err := ctr.OvertimeRepo.SelectOvertimeRequests(userProfile.OrganizationID, params)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "A", 20)
	_ = f.SetColWidth("Sheet1", "B", "B", 30)
	_ = f.SetColWidth("Sheet1", "C", "F", 15)
	_ = f.SetColWidth("Sheet1", "G", "H", 20)
	_ = f.SetColWidth("Sheet1", "I", "I", 15)
	_ = f.SetColWidth("Sheet1", "J", "J", 30)
	_ = f.SetColWidth("Sheet1", "K", "M", 15)

	titleStyle, _ := f.NewStyle(`{
		"font":{"bold":true, "size":16},
		"alignment":{"horizontal":"center", "vertical":"center"}
	}`)
	contentStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center", "vertical":"center"}}`)

	var categoriesByLanguage map[string]string
	var dayOfWeekByLanguage map[string]string
	var overtimeTypeLanguage map[int]string
	var mapStatusOvertimeType map[int]string

	if userProfile.LanguageId == cf.EnLanguageId {
		categoriesByLanguage = cf.EnOvertimeCategories
		dayOfWeekByLanguage = cf.EnWeekDay
		overtimeTypeLanguage = cf.MapOvertimeType
		mapStatusOvertimeType = cf.MapStatusOvertimeType
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categoriesByLanguage = cf.VnOvertimeCategories
		dayOfWeekByLanguage = cf.VnWeekDay
		overtimeTypeLanguage = cf.VnMapOvertimeType
		mapStatusOvertimeType = cf.VnMapStatusOvertimeType
	} else {
		categoriesByLanguage = cf.JpOvertimeCategories
		dayOfWeekByLanguage = cf.JpWeekDay
		overtimeTypeLanguage = cf.JpMapOvertimeType
		mapStatusOvertimeType = cf.JpMapStatusOvertimeType
	}

	categories := map[string]string{
		"A1": categoriesByLanguage["Employee Id"],
		"B1": categoriesByLanguage["Full Name"],
		"C1": categoriesByLanguage["Branch"],
		"D1": categoriesByLanguage["Project"],
		"E1": categoriesByLanguage["Date"],
		"F1": categoriesByLanguage["Weekday"],
		"G1": categoriesByLanguage["Range Time"],
		"H1": categoriesByLanguage["Working Time"],
		"I1": categoriesByLanguage["Weight"],
		"J1": categoriesByLanguage["Total Working Time"],
		"K1": categoriesByLanguage["Type"],
		"L1": categoriesByLanguage["Status"],
		"M1": categoriesByLanguage["Note"],
	}
	for k, v := range categories {
		_ = f.SetCellValue("Sheet1", k, v)
	}
	_ = f.SetCellStyle("Sheet1", "A1", "M1", titleStyle)
	_ = f.SetColStyle("Sheet1", "A", contentStyle)
	_ = f.SetColStyle("Sheet1", "C", contentStyle)
	_ = f.SetColStyle("Sheet1", "D", contentStyle)
	_ = f.SetColStyle("Sheet1", "E", contentStyle)
	_ = f.SetColStyle("Sheet1", "F", contentStyle)
	_ = f.SetColStyle("Sheet1", "G", contentStyle)
	_ = f.SetColStyle("Sheet1", "H", contentStyle)
	_ = f.SetColStyle("Sheet1", "I", contentStyle)
	_ = f.SetColStyle("Sheet1", "J", contentStyle)
	_ = f.SetColStyle("Sheet1", "K", contentStyle)
	_ = f.SetColStyle("Sheet1", "L", contentStyle)

	idx := 2
	for i, record := range otRecords {
		pos := idx + i
		holidayDates := ctr.getHolidayCurrentYear(userProfile.OrganizationID, record.DatetimeOvertimeFrom.Year())
		cld := calendar.NewCalendar(holidayDates)
		actualHour, hour, weight := ctr.calculateActualHourOvertime(
			cld,
			userProfile.OrganizationID,
			record.DatetimeOvertimeFrom,
			record.DatetimeOvertimeTo,
			record.WorkAtNoon,
		)

		values := map[string]interface{}{
			"A" + strconv.Itoa(pos): record.EmployeeId,
			"B" + strconv.Itoa(pos): record.FullName,
			"C" + strconv.Itoa(pos): record.Branch,
			"D" + strconv.Itoa(pos): record.ProjectName,
			"E" + strconv.Itoa(pos): record.DatetimeOvertimeFrom.Format(cf.FormatDateDisplay),
			"F" + strconv.Itoa(pos): dayOfWeekByLanguage[strconv.Itoa(int(record.DatetimeOvertimeFrom.Weekday()))],
			"G" + strconv.Itoa(pos): utils.ConvertTwoChar(record.HourFrom) + ":" + utils.ConvertTwoChar(record.MinuteFrom) +
				" - " + utils.ConvertTwoChar(record.HourTo) + ":" + utils.ConvertTwoChar(record.MinuteTo),
			"H" + strconv.Itoa(pos): hour,
			"I" + strconv.Itoa(pos): weight,
			"J" + strconv.Itoa(pos): actualHour,
			"K" + strconv.Itoa(pos): overtimeTypeLanguage[record.OvertimeType],
			"L" + strconv.Itoa(pos): mapStatusOvertimeType[record.Status],
		}

		for k, v := range values {
			_ = f.SetCellValue("Sheet1", k, v)
		}
	}

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}

func (ctr *Controller) CreateOvertimeWeight(c echo.Context) error {
	params := new(param.CreateOvertimeWeightParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil ||
		!valid.IsPositive(params.NormalDayWeight) ||
		!valid.IsPositive(params.WeekendWeight) ||
		!valid.IsPositive(params.HolidayWeight) {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.OvertimeRepo.CountOvertimeWeightByField("organization_id", userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if count > 0 {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Overtime weight already exist",
		})
	}

	err = ctr.OvertimeRepo.InsertOvertimeWeight(userProfile.OrganizationID, params, userProfile.Organization.SettingStep, ctr.OrgRepo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create overtime weight successful",
	})
}

func (ctr *Controller) EditOvertimeWeight(c echo.Context) error {
	params := new(param.EditOvertimeWeightParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if (params.NormalDayWeight != 0 && !valid.IsPositive(params.NormalDayWeight)) ||
		(params.WeekendWeight != 0 && !valid.IsPositive(params.WeekendWeight)) ||
		(params.HolidayWeight != 0 && !valid.IsPositive(params.HolidayWeight)) {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	count, err := ctr.OvertimeRepo.CountOvertimeWeightByField("id", params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if count == 0 {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Overtime weight does not yet exist",
		})
	}

	err = ctr.OvertimeRepo.UpdateOvertimeWeight(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit overtime weight successful",
	})
}

func (ctr *Controller) GetOvertimeWeight(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	record, err := ctr.OvertimeRepo.SelectOvertimeWeightByOrganizationId(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Overtime weight is empty",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	dataResponse := map[string]interface{}{
		"id":                record.ID,
		"normal_day_weight": record.NormalDayWeight,
		"weekend_weight":    record.WeekendWeight,
		"holiday_weight":    record.HolidayWeight,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get overtime weight successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) calculateActualHourOvertime(
	c *calendar.Calendar,
	organizationId int,
	from time.Time,
	to time.Time,
	workAtNoon int,
) (float64, float64, float64) {
	overtimeWeight, err := ctr.OvertimeRepo.SelectOvertimeWeightByOrganizationId(organizationId)
	if err != nil {
		return -1, -1, 0
	}

	actualHour, hour, weight := calendar.CalculateHourBonusOvertime(c, from, to, overtimeWeight, workAtNoon)
	return actualHour, hour, weight
}

func (ctr *Controller) getHolidayCurrentYear(organizationId int, year int) []time.Time {
	records, err := ctr.HolidayRepo.SelectHolidays(organizationId, year, "holiday_date")
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		panic(err)
	}

	var holidayDates []time.Time
	for _, record := range records {
		holidayDates = append(holidayDates, record.HolidayDate)
	}

	return holidayDates
}
