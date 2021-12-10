package leave

import (
	"net/http"
	"strconv"

	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"

	valid "github.com/asaskevich/govalidator"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	cr "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cron"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	ex "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/excel"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

// LvController : Struct controller
type LvController struct {
	cm.BaseController
	email.SMTPGoMail
	cr.EtrCron
	afb.FirebaseCloudMessage

	LeaveRepo        rp.LeaveRepository
	UserRepo         rp.UserRepository
	BranchRepo       rp.BranchRepository
	OrgRepo          rp.OrgRepository
	HolidayRepo      rp.HolidayRepository
	NotificationRepo rp.NotificationRepository
	UserProjectRepo  rp.UserProjectRepository
	FcmTokenRepo     rp.FcmTokenRepository
	cloud            gc.StorageUtility
}

// NewLeaveController : Init controller
func NewLeaveController(
	logger echo.Logger,
	leaveRepo rp.LeaveRepository,
	userRepo rp.UserRepository,
	branchRepo rp.BranchRepository,
	orgRepo rp.OrgRepository,
	holidayRepo rp.HolidayRepository,
	notificationRepo rp.NotificationRepository,
	userProjectRepo rp.UserProjectRepository,
	fcmTokenRepo rp.FcmTokenRepository,
	cloud gc.StorageUtility,
) (ctr *LvController) {
	ctr = &LvController{
		cm.BaseController{}, email.SMTPGoMail{}, cr.EtrCron{}, afb.FirebaseCloudMessage{},
		leaveRepo, userRepo, branchRepo, orgRepo,
		holidayRepo, notificationRepo, userProjectRepo, fcmTokenRepo, cloud,
	}
	ctr.Init(logger)
	ctr.InitCron("Asia/Ho_Chi_Minh")
	ctr.InitFcm()
	return
}

// CreateLeaveRequest : Create user leave request
func (ctr *LvController) CreateLeaveRequest(c echo.Context) error {
	createLeaveRequestParams := new(param.CreateLeaveRequestParams)
	if err := c.Bind(createLeaveRequestParams); err != nil {
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

	if org.Email != "" && org.EmailPassword != "" {
		ctr.InitSmtp(org.Email, org.EmailPassword)
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
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

	users := make(map[int]string)
	for i := 0; i < len(userRecords); i++ {
		users[userRecords[i].UserID] = userRecords[i].FullName
	}

	usersIdProject, err := ctr.UserProjectRepo.SelectUserIdsJoinProjectsWithUserId(userProfile.OrganizationID, userProfile.UserProfile.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	usersIdGmAndManager, err := ctr.UserRepo.SelectIdsOfGMAndManager(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	uniqueUsersId := utils.AppendUniqueSlice(usersIdProject, usersIdGmAndManager)
	var hourReq float64
	for i, createLeaveRequestParam := range createLeaveRequestParams.LeaveRequest {
		createLeaveRequestParam.OrgID = userProfile.OrganizationID
		createLeaveRequestParam.CreatedBy, createLeaveRequestParam.UpdatedBy = userProfile.UserProfile.UserID, userProfile.UserProfile.UserID

		_, err := valid.ValidateStruct(createLeaveRequestParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i,
			})
		}

		if userProfile.RoleID == cf.UserRoleID && createLeaveRequestParam.UserID != userProfile.UserProfile.UserID {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You can't create leave request for another user.",
				Data:    i,
			})
		}

		if (createLeaveRequestParam.LeaveRequestTypeID == cf.LateForWork ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.LeaveEarly ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.GoOutside) && createLeaveRequestParam.DatetimeLeaveTo == "" {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "To date is required.",
				Data:    i,
			})
		}

		timestampFrom := calendar.ParseTime(cf.FormatDateNoSec, createLeaveRequestParam.DatetimeLeaveFrom)
		duration, _ := time.ParseDuration("12h0m0s")
		if calendar.ParseTime(cf.FormatDateNoSec, time.Now().Format(cf.FormatDateNoSec)).Sub(timestampFrom) > duration {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You can't create leave request.",
				Data:    i,
			})
		}

		timestampTo := calendar.ParseTime(cf.FormatDateNoSec, createLeaveRequestParam.DatetimeLeaveTo)
		if timestampTo.Sub(timestampFrom) < 0 {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "From date must be less than To date.",
				Data:    i,
			})
		}

		if (createLeaveRequestParam.LeaveRequestTypeID == cf.MorningOff ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.AfternoonOff ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.LateForWork ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.LeaveEarly ||
			createLeaveRequestParam.LeaveRequestTypeID == cf.GoOutside) && !utils.CompareEqualDate(timestampTo, timestampFrom) {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "From date & To date not same.",
				Data:    i,
			})
		}

		lunchBreakStart := calendar.ParseTime(cf.FormatDateNoSec, timestampTo.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchStart)
		if createLeaveRequestParam.LeaveRequestTypeID == cf.LateForWork && lunchBreakStart.Sub(timestampFrom) <= 0 {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Time from must be less 12:00.",
				Data:    i,
			})
		}

		lunchBreakEnd := calendar.ParseTime(cf.FormatDateNoSec, timestampTo.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchEnd)
		if createLeaveRequestParam.LeaveRequestTypeID == cf.LeaveEarly && timestampTo.Sub(lunchBreakEnd) <= 0 {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Time to must be greater 13:30.",
				Data:    i,
			})
		}

		Id, body, link, hour, err := ctr.LeaveRepo.InsertLeaveRequest(&createLeaveRequestParam, ctr.HolidayRepo, ctr.UserRepo, ctr.NotificationRepo, uniqueUsersId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
				Data:    i,
			})
		}
		hourReq = hour

		registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(uniqueUsersId, createLeaveRequestParam.UserID)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
				Data:    i,
			})
		}

		if len(registrationTokens) > 0 {
			for _, token := range registrationTokens {
				err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
				if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
					"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
					_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
				}
			}
		}

		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{cf.NoticeEmail}
		sampleData.Content = createLeaveRequestParam.EmailContent
		if err := ctr.SendMail(createLeaveRequestParam.EmailTitle, sampleData, cf.LeaveRequestTemplate); err != nil {
			ctr.Logger.Error(err)
		}

		startDate, endDate := strings.Split(createLeaveRequestParam.DatetimeLeaveFrom, " ")[0], strings.Split(createLeaveRequestParam.DatetimeLeaveTo, " ")[0]
		startTime, endTime := strings.Split(createLeaveRequestParam.DatetimeLeaveFrom, " ")[1], strings.Split(createLeaveRequestParam.DatetimeLeaveTo, " ")[1]
		var start, end string
		switch createLeaveRequestParam.LeaveRequestTypeID {
		case cf.FullDayOff, cf.MorningOff, cf.AfternoonOff:
			start = startDate
			end = calendar.ParseTime(cf.FormatDateDatabase, endDate).AddDate(0, 0, 1).Format(cf.FormatDateDatabase)
		case cf.LateForWork, cf.LeaveEarly, cf.GoOutside:
			start = startDate + "T" + startTime + ":00+07:00"
			end = endDate + "T" + endTime + ":00+07:00"
		default:
			hourOfDay, _ := time.ParseDuration("8h0m0s")
			if 0 < timestampTo.Sub(timestampFrom) && timestampTo.Sub(timestampFrom) < hourOfDay {
				start = startDate + "T" + startTime + ":00+07:00"
				end = endDate + "T" + endTime + ":00+07:00"
			} else {
				start = startDate
				end = calendar.ParseTime(cf.FormatDateDatabase, endDate).AddDate(0, 0, 1).Format(cf.FormatDateDatabase)
			}
		}

		event := calendar.AddLeaveEvent(
			createLeaveRequestParam.LeaveRequestTypeID,
			users[createLeaveRequestParam.UserID]+" - "+cf.LeaveRequestJpTypes[createLeaveRequestParam.LeaveRequestTypeID],
			createLeaveRequestParam.Reason,
			start,
			end,
		)

		err = ctr.LeaveRepo.UpdateLeaveRequest(Id, event.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	}

	returnObj := map[string]float64{
		"hour": hourReq,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create leave request successfully.",
		Data:    returnObj,
	})
}

// CreateLeaveBonus : Create user leave bonus
func (ctr *LvController) CreateLeaveBonus(c echo.Context) error {
	createLeaveBonusParams := new(param.CreateLeaveBonusParams)
	if err := c.Bind(createLeaveBonusParams); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	for i, createLeaveBonusParam := range createLeaveBonusParams.LeaveBonus {
		_, err := valid.ValidateStruct(createLeaveBonusParam)
		if err != nil {
			ctr.Logger.Error(err)
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i,
			})
		}

		if createLeaveBonusParam.LeaveBonusTypeID < cf.ClearLeave && createLeaveBonusParam.Hour <= 0 {
			return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i,
			})
		}
	}

	userProfile := c.Get("user_profile").(m.User)
	if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
		userProfile.OrganizationID,
		userProfile.UserProfile.UserID,
		&createLeaveBonusParams.LeaveBonus,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create leave bonus successfully.",
	})
}

// GetLeaveInfo : Get used, remaining leave day of user
func (ctr *LvController) GetLeaveInfo(c echo.Context) error {
	leaveStatusParams := new(param.LeaveStatusParams)
	if err := c.Bind(leaveStatusParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	leaveStatusParams.OrgID = userProfile.OrganizationID

	_, err := valid.ValidateStruct(leaveStatusParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	user, err := ctr.UserRepo.GetUserProfileExpand(leaveStatusParams.UserID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User profile not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
			Data:    err,
		})
	}

	var base64Img []byte = nil
	if user.Avatar != "" {
		base64Img, err = ctr.cloud.GetFileByFileName(user.Avatar, cf.AvatarFolderGCS)

		if err != nil {
			ctr.Logger.Error(err)
			base64Img = nil
		}
	}

	userInfo := map[string]interface{}{
		"email":        user.Email,
		"phone_number": user.PhoneNumber,
		"full_name":    user.FirstName + " " + user.LastName,
		"avatar":       base64Img,
	}

	dayUsed, dayBonus, dayRemaining, dayRemainingPrevious := ctr.LeaveRepo.GetLeaveDayStatus(
		leaveStatusParams.OrgID,
		leaveStatusParams.UserID,
		leaveStatusParams.Year,
	)

	holidayDates := ctr.getHolidaysOfYear(leaveStatusParams.OrgID, leaveStatusParams.Year)
	var holidays []string
	holidaysVN := calendar.GetHolidays(holidayDates)

	for _, holiday := range holidaysVN {
		hld := strconv.Itoa(holiday.Year) + "-" + strconv.Itoa(int(holiday.Month)) + "-" + strconv.Itoa(holiday.Day)
		holidays = append(holidays, hld)
	}

	dataResponse := map[string]interface{}{
		"user_info":              userInfo,
		"leave_request_types":    cf.LeaveRequestTypes,
		"leave_bonus_types":      cf.LeaveBonusTypes,
		"holidays":               holidays,
		"day_used":               dayUsed,
		"day_bonus":              dayBonus,
		"day_remaining":          dayRemaining,
		"day_remaining_previous": dayRemainingPrevious,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get leave info successfully.",
		Data:    dataResponse,
	})
}

// GetLeaveHistory : Get used, remaining leave day of user
func (ctr *LvController) GetLeaveHistory(c echo.Context) error {
	leaveHistoryParams := new(param.LeaveHistoryParams)

	if err := c.Bind(leaveHistoryParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	leaveHistoryParams.RowPerPage = 10

	leaveHistoryList, totalLeaveHistory, err := ctr.LeaveRepo.SearchUserLeave(userProfile.OrganizationID, leaveHistoryParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get leave histories failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	leaveHistoryResponse := []map[string]interface{}{}
	if totalLeaveHistory > 0 {
		for _, userItem := range leaveHistoryList {
			var base64Img []byte = nil
			if userItem.UserProfile.Avatar != "" {
				base64Img, err = ctr.cloud.GetFileByFileName(userItem.UserProfile.Avatar, cf.AvatarFolderGCS)

				if err != nil {
					ctr.Logger.Error(err)
					base64Img = nil
				}
			}

			arrLeaveRequest := []map[string]interface{}{}
			for _, leaveItem := range userItem.UserLeaveRequest {
				leaveObj := map[string]interface{}{
					"leave_request_type_id":    leaveItem.LeaveRequestTypeID,
					"subtract_day_off_type_id": leaveItem.SubtractDayOffTypeID,
					"date_time_leave_from":     leaveItem.DatetimeLeaveFrom.Format(cf.FormatDateNoSec),
					"date_time_leave_to":       leaveItem.DatetimeLeaveTo.Format(cf.FormatDateNoSec),
				}

				arrLeaveRequest = append(arrLeaveRequest, leaveObj)
			}

			leaveHistoryObj := map[string]interface{}{
				"id":            userItem.ID,
				"first_name":    userItem.UserProfile.FirstName,
				"last_name":     userItem.UserProfile.LastName,
				"avatar":        base64Img,
				"leave_request": arrLeaveRequest,
			}
			leaveHistoryResponse = append(leaveHistoryResponse, leaveHistoryObj)
		}
	}

	userList, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userListRes := make(map[int]string)
	for i := 0; i < len(userList); i++ {
		userListRes[userList[i].UserID] = userList[i].FullName
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"user_list":  userListRes,
			"user_leave": leaveHistoryResponse,
			"pagination": map[string]interface{}{
				"current_page": leaveHistoryParams.CurrentPage,
				"total_row":    totalLeaveHistory,
				"row_per_page": leaveHistoryParams.RowPerPage,
			},
			"subtract_day_off_types": cf.SubtractDayOffTypes,
			"leave_request_types":    cf.LeaveRequestTypes,
		},
	})
}

// GetLeaveRequests : Get leave requests of users
func (ctr *LvController) GetLeaveRequests(c echo.Context) error {
	leaveRequestListParams := new(param.LeaveRequestListParams)
	leaveRequestListParams.RowPerPage = 10

	if err := c.Bind(leaveRequestListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	columns := []string{"ulr.id", "ulr.user_id", "ulr.leave_request_type_id", "u.email", "up.avatar", "ulr.datetime_leave_from", "ulr.datetime_leave_to", "ulr.email_content", "ulr.email_title", "ulr.reason"}
	leaveRequestRecords, totalRow, err := ctr.LeaveRepo.LeaveRequests(
		userProfile.OrganizationID,
		"ulr.datetime_leave_from DESC",
		leaveRequestListParams,
		true,
		columns...,
	)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get leave requests failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if leaveRequestListParams.RowPerPage == 0 {
		leaveRequestListParams.CurrentPage = 1
		leaveRequestListParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": leaveRequestListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": leaveRequestListParams.RowPerPage,
	}

	var leaveRequestsResponse []map[string]interface{}

	for _, leaveRequest := range leaveRequestRecords {
		var base64Img []byte = nil
		if leaveRequest.Avatar != "" {
			base64Img, err = ctr.cloud.GetFileByFileName(leaveRequest.Avatar, cf.AvatarFolderGCS)

			if err != nil {
				ctr.Logger.Error(err)
				base64Img = nil
			}
		}
		leave := map[string]interface{}{
			"id":                    leaveRequest.ID,
			"user_id":               leaveRequest.UserID,
			"leave_request_type_id": leaveRequest.LeaveRequestTypeID,
			"email":                 leaveRequest.Email,
			"avatar":                base64Img,
			"full_name":             leaveRequest.FullName,
			"email_content":         leaveRequest.EmailContent,
			"email_title":           leaveRequest.EmailTitle,
			"reason":                leaveRequest.Reason,
			"datetime_leave_from":   leaveRequest.DatetimeLeaveFrom,
			"isShow":                false,
		}

		switch leaveRequest.LeaveRequestTypeID {
		case cf.FullDayOff, cf.WorkAtHome, cf.BusinessTrip:
			leave["datetime_leave_from"] = leaveRequest.DatetimeLeaveFrom.Format(cf.FormatDateDisplay)
			if !utils.CompareEqualDate(leaveRequest.DatetimeLeaveTo, leaveRequest.DatetimeLeaveFrom) {
				leave["datetime_leave_to"] = leaveRequest.DatetimeLeaveTo.Format(cf.FormatDateDisplay)
			}
		case cf.MorningOff, cf.AfternoonOff:
			leave["datetime_leave_from"] = leaveRequest.DatetimeLeaveFrom.Format(cf.FormatDateDisplay)
		default:
			hourTo := utils.ConvertTwoChar(leaveRequest.HourTo)
			minuteTo := utils.ConvertTwoChar(leaveRequest.MinuteTo)
			leave["datetime_leave_from"] = leaveRequest.DatetimeLeaveFrom.Format(cf.FormatTimeDisplay) + "-" + hourTo + ":" + minuteTo
		}

		leaveRequestsResponse = append(leaveRequestsResponse, leave)
	}

	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
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

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination":          pagination,
			"branch_select_box":   branches,
			"leave_requests":      leaveRequestsResponse,
			"user_list":           userList,
			"leave_request_types": cf.LeaveRequestTypes,
		},
	})
}

func (ctr *LvController) getLeaveHistoriesByUser(organizationId int, leaveRequest param.LeaveHistoryRecords) param.LeaveHistoryByUser {
	leaveHistoryByUser := param.LeaveHistoryByUser{}
	leaveHistoryByUser.LeaveRequestTypeID = leaveRequest.LeaveRequestTypeID
	leaveHistoryByUser.SubtractDayOffTypeID = leaveRequest.SubtractDayOffTypeID

	holidayDates := ctr.getHolidaysOfYear(organizationId, leaveRequest.DatetimeLeaveFrom.Year())
	cal := calendar.NewCalendar(holidayDates)

	switch leaveRequest.LeaveRequestTypeID {
	case cf.FullDayOff, cf.WorkAtHome, cf.BusinessTrip:
		if utils.CompareEqualDate(leaveRequest.DatetimeLeaveTo, leaveRequest.DatetimeLeaveFrom) {
			if !cal.IsHoliday(leaveRequest.DatetimeLeaveFrom) && !calendar.IsWeekend(leaveRequest.DatetimeLeaveFrom) {
				leaveHistoryByUser.LeaveDates = append(
					leaveHistoryByUser.LeaveDates,
					leaveRequest.DatetimeLeaveFrom.Format(cf.FormatDateDisplay),
				)
			}
		} else {
			var dates []time.Time
			diff := int(leaveRequest.DatetimeLeaveTo.Sub(leaveRequest.DatetimeLeaveFrom).Hours()/24) + 1
			for i := 0; i < diff; i++ {
				dates = append(dates, leaveRequest.DatetimeLeaveFrom.AddDate(0, 0, i))
			}

			for _, date := range dates {
				holidayDates := ctr.getHolidaysOfYear(organizationId, date.Year())
				cal2 := calendar.NewCalendar(holidayDates)
				if !cal2.IsHoliday(date) && !calendar.IsWeekend(date) {
					leaveHistoryByUser.LeaveDates = append(
						leaveHistoryByUser.LeaveDates,
						date.Format(cf.FormatDateDisplay),
					)
				}
			}
		}
	case cf.MorningOff, cf.AfternoonOff:
		if !cal.IsHoliday(leaveRequest.DatetimeLeaveFrom) && !calendar.IsWeekend(leaveRequest.DatetimeLeaveFrom) {
			leaveHistoryByUser.LeaveDates = append(
				leaveHistoryByUser.LeaveDates,
				leaveRequest.DatetimeLeaveFrom.Format(cf.FormatDateDisplay),
			)
		}
	default:
		if !cal.IsHoliday(leaveRequest.DatetimeLeaveFrom) && !calendar.IsWeekend(leaveRequest.DatetimeLeaveFrom) {
			hourTo := utils.ConvertTwoChar(leaveRequest.HourTo)
			minuteTo := utils.ConvertTwoChar(leaveRequest.MinuteTo)

			leaveDate := leaveRequest.DatetimeLeaveFrom.Format(cf.FormatTimeDisplay) + "-" + hourTo + ":" + minuteTo
			if leaveRequest.SubtractDayOffTypeID == cf.ExtraWork {
				leaveDate += " (ExtraWork)"
			}
			if leaveRequest.SubtractDayOffTypeID == cf.Event {
				leaveDate += " (Event)"
			}

			leaveHistoryByUser.LeaveDates = append(
				leaveHistoryByUser.LeaveDates,
				leaveDate,
			)
		}
	}

	return leaveHistoryByUser
}

// RemoveLeaveRequest : Remove leave request of user
func (ctr *LvController) RemoveLeaveRequest(c echo.Context) error {
	removeLeaveParams := new(param.RemoveLeaveParams)

	if err := c.Bind(removeLeaveParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(removeLeaveParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	leaveRequest, err := ctr.LeaveRepo.SelectLeaveRequestById(removeLeaveParams.LeaveID)
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
			Data:    err,
		})
	}

	validDateRecords, err := ctr.LeaveRepo.SelectValidDateOfUser(leaveRequest.OrganizationID, leaveRequest.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	hourLeaveTemp := leaveRequest.Hour
	if len(validDateRecords) > 0 {
		for _, validDate := range validDateRecords {
			var hourUsed float64 = validDate.Hour - validDate.HourRemaining
			if hourLeaveTemp < hourUsed {
				ctr.LeaveRepo.UpdateHourRemainingLeaveByID(validDate.Id, validDate.HourRemaining+hourLeaveTemp)
			} else {
				ctr.LeaveRepo.UpdateHourRemainingLeaveByID(validDate.Id, validDate.HourRemaining+hourUsed)
			}
			hourLeaveTemp = hourLeaveTemp - hourUsed
			if hourLeaveTemp <= 0 {
				break
			}
		}
	}

	err = ctr.LeaveRepo.RemoveLeave(removeLeaveParams.LeaveID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if leaveRequest.CalendarEventId != "" {
		calendar.RemoveLeaveEvent(leaveRequest.CalendarEventId)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove leave request successfully.",
	})
}

// GetLeaveInfoAllUser : Get leave info of all user
func (ctr *LvController) GetLeaveInfoAllUser(c echo.Context) error {
	allUserNameAndCountParams := new(param.AllUserNameAndCountParams)
	allUserNameAndCountParams.RowPerPage = 10

	if err := c.Bind(allUserNameAndCountParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	allUserNameAndCountParams.OrgID = userProfile.OrganizationID

	_, err := valid.ValidateStruct(allUserNameAndCountParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	users, totalRow, err := ctr.UserRepo.GetAllUserNameAndCountByOrgID(allUserNameAndCountParams)

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

	var userLeaveInfoResponse []map[string]interface{}
	for _, user := range users {
		var base64Img []byte = nil

		if user.Avatar != "" {
			base64Img, err = ctr.cloud.GetFileByFileName(user.Avatar, cf.AvatarFolderGCS)

			if err != nil {
				ctr.Logger.Error(err)
				base64Img = nil
			}
		}
		userLeaveInfo := map[string]interface{}{
			"user_id":   user.UserID,
			"full_name": user.FullName,
			"branch":    user.Branch,
			"email":     user.Email,
			"avatar":    base64Img,
		}

		dayUsed, _, dayRemaining, _ := ctr.LeaveRepo.GetLeaveDayStatus(allUserNameAndCountParams.OrgID, user.UserID, time.Now().Year())
		userLeaveInfo["day_used"] = dayUsed
		userLeaveInfo["day_remaining"] = dayRemaining

		userLeaveInfoResponse = append(userLeaveInfoResponse, userLeaveInfo)
	}

	allUsers, err := ctr.UserRepo.GetAllUserNameByOrgID(allUserNameAndCountParams.OrgID)

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

	usersBox := make(map[int]string)

	for i := 0; i < len(allUsers); i++ {
		usersBox[allUsers[i].UserID] = allUsers[i].FullName
	}

	if allUserNameAndCountParams.RowPerPage == 0 {
		allUserNameAndCountParams.CurrentPage = 1
		allUserNameAndCountParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": allUserNameAndCountParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": allUserNameAndCountParams.RowPerPage,
	}

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

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination":        pagination,
			"branch_select_box": branches,
			"user_list":         usersBox,
			"users_leave_info":  userLeaveInfoResponse,
		},
	})
}

// CronLeaveBonus : Cron annual leave bonus with run year-01-01 00:00:00 and clear old leave with run year-04-01 00:00:00
func (ctr *LvController) CronLeaveBonus(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	_, err = ctr.AddFuncCron("* * 1 1 *", "Annual leave bonus cron", func() {
		var leaveBonusParams []param.LeaveBonus
		t := time.Now()
		firstday := time.Date(t.Year()+1, time.Month(userProfile.Organization.ExpirationResetDayOff+1), 1, 0, 0, 0, 0, time.Local)
		for _, user := range users {
			leaveBonusParam := new(param.LeaveBonus)
			leaveBonusParam.UserID = user.UserID
			leaveBonusParam.LeaveBonusTypeID = cf.AnnualLeave
			leaveBonusParam.YearBelong = time.Now().Year()
			leaveBonusParam.Reason = "Annual Leave"
			leaveBonusParam.Hour = 96
			leaveBonusParam.ExpireBonusLeaveDate = firstday.Format("2006-01-02 00:04:05")
			leaveBonusParam.HourRemaining = 96

			leaveBonusParams = append(leaveBonusParams, *leaveBonusParam)
		}

		if len(leaveBonusParams) > 0 {
			if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
				userProfile.OrganizationID,
				userProfile.UserProfile.UserID,
				&leaveBonusParams,
			); err != nil {
				ctr.Logger.Error(err)
			}
		}
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	_, err = ctr.AddFuncCron("* * 1 * *", "Seniority leave bonus cron", func() {
		t := time.Now()
		firstday := time.Date(t.Year()+1, time.Month(userProfile.Organization.ExpirationResetDayOff+1), 1, 0, 0, 0, 0, time.Local)

		var seniorityLeaveBonus []param.LeaveBonus
		for _, user := range users {
			if !user.CompanyJoinedDate.IsZero() && user.CompanyJoinedDate.Month() == t.Month() {
				leaveBonusParam := new(param.LeaveBonus)
				leaveBonusParam.UserID = user.UserID
				leaveBonusParam.LeaveBonusTypeID = cf.SeniorityLeave
				leaveBonusParam.YearBelong = time.Now().Year()
				leaveBonusParam.Reason = "Seniority Leave"
				workedMonths := monthsCountSince(user.CompanyJoinedDate)
				monthCompanyJoined := user.CompanyJoinedDate.Month()

				switch {
				case workedMonths < 12:
					break
				case workedMonths >= 12 && workedMonths < 24:
					leaveBonusParam.Hour = float64(8*(12-monthCompanyJoined)/12)
					leaveBonusParam.HourRemaining = float64(8*(12-monthCompanyJoined)/12)
					break
				case workedMonths >= 24 && workedMonths < 36:
					leaveBonusParam.Hour = 8
					leaveBonusParam.HourRemaining = 8
					break
				case workedMonths >= 36 && workedMonths < 48:
					leaveBonusParam.Hour = float64(8+(8*(12-monthCompanyJoined)/12))
					leaveBonusParam.HourRemaining = float64(8+(8*(12-monthCompanyJoined)/12))
					break
				case workedMonths >= 48 && workedMonths < 60:
					leaveBonusParam.Hour = 16
					leaveBonusParam.HourRemaining = 16
					break
				case workedMonths >= 60 && workedMonths < 72:
					leaveBonusParam.Hour = float64(16+(8*(12-monthCompanyJoined)/12))
					leaveBonusParam.HourRemaining = float64(16+(8*(12-monthCompanyJoined)/12))
					break
				case workedMonths >= 72:
					leaveBonusParam.Hour = 24
					leaveBonusParam.HourRemaining = 24
					break
				}

				leaveBonusParam.ExpireBonusLeaveDate = firstday.Format("2006-01-02 00:04:05")
				seniorityLeaveBonus = append(seniorityLeaveBonus, *leaveBonusParam)
			}
		}

		if len(seniorityLeaveBonus) > 0 {
			if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
				userProfile.OrganizationID,
				userProfile.UserProfile.UserID,
				&seniorityLeaveBonus,
			); err != nil {
				ctr.Logger.Error(err)
			}
		}
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var spec string
	if userProfile.Organization.ExpirationResetDayOff < 12 {
		spec = "5 9 2 " + strconv.Itoa(userProfile.Organization.ExpirationResetDayOff+1) + " *"
	} else {
		spec = "0 0 1 1 *"
	}
	_, err = ctr.AddFuncCron(spec, "Clear old leave cron", func() {
		previousYear := time.Now().Year() - 1
		var leaveBonusParams []param.LeaveBonus

		for _, user := range users {
			hourUsedOld, err := ctr.LeaveRepo.CountHourUsed(userProfile.OrganizationID, user.UserID, previousYear)
			if err != nil {
				return
			}

			hourBonusOld, err := ctr.LeaveRepo.CountHourBonus(userProfile.OrganizationID, user.UserID, previousYear)
			if err != nil {
				return
			}

			t := time.Now()
			first := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
			hoursUpOld, err := ctr.LeaveRepo.CountHoursUpToExpirationDate(userProfile.OrganizationID, user.UserID, t.Format(cf.FormatDate), first.Format(cf.FormatDate))
			if err != nil {
				return
			}

			hourRemainingOld := hourBonusOld - hourUsedOld - hoursUpOld
			if hourRemainingOld > 0 {
				leaveBonusParam := new(param.LeaveBonus)
				leaveBonusParam.UserID = user.UserID
				leaveBonusParam.LeaveBonusTypeID = cf.AnnualLeave
				leaveBonusParam.YearBelong = previousYear
				leaveBonusParam.Reason = "Clear old leave"
				leaveBonusParam.Hour = -hourRemainingOld
				leaveBonusParam.HourRemaining = -hourRemainingOld

				leaveBonusParams = append(leaveBonusParams, *leaveBonusParam)
			}
		}

		if len(leaveBonusParams) > 0 {
			if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
				userProfile.OrganizationID,
				userProfile.UserProfile.UserID,
				&leaveBonusParams,
			); err != nil {
				ctr.Logger.Error(err)
			}
		}
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	_, err = ctr.AddFuncCron("0 0 1 * *", "Add extra day at the beginning of the month when running out of day off", func() {
		currentYear := time.Now().Year()
		var leaveBonusParams []param.LeaveBonus

		for _, user := range users {
			hourUsedCurrent, err := ctr.LeaveRepo.CountHourUsed(userProfile.OrganizationID, user.UserID, currentYear)
			if err != nil {
				return
			}

			hourBonusCurrent, err := ctr.LeaveRepo.CountHourBonus(userProfile.OrganizationID, user.UserID, currentYear)
			if err != nil {
				return
			}

			hourRemainingOld := hourBonusCurrent - hourUsedCurrent
			if hourRemainingOld < 0 {
				leaveBonusParam := new(param.LeaveBonus)
				leaveBonusParam.UserID = user.UserID
				leaveBonusParam.LeaveBonusTypeID = cf.ClearLeave
				leaveBonusParam.YearBelong = currentYear
				leaveBonusParam.Reason = "Add extra day at the beginning of the month when running out of day off"
				leaveBonusParam.Hour = -hourRemainingOld

				leaveBonusParams = append(leaveBonusParams, *leaveBonusParam)
			}
		}

		if len(leaveBonusParams) > 0 {
			if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
				userProfile.OrganizationID,
				userProfile.UserProfile.UserID,
				&leaveBonusParams,
			); err != nil {
				ctr.Logger.Error(err)
			}
		}
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create cron leave successfully.",
	})
}

// CronLeaveStart : Start leave cron
func (ctr *LvController) CronLeaveStart(c echo.Context) error {
	ctr.StartCron()
	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Start cron leave successfully.",
	})
}

// CronLeaveStop : Stop leave cron
func (ctr *LvController) CronLeaveStop(c echo.Context) error {
	ctr.StopCron()
	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Stop cron leave successfully.",
	})
}

// GetEntriesLeave : Get entries leave
func (ctr *LvController) GetEntriesLeave(c echo.Context) error {
	entries := ctr.GetEntries()
	var entriesResponse []map[string]interface{}

	for _, entry := range entries {
		itemDataResponse := map[string]interface{}{
			"id":   entry.ID,
			"name": entry.Name,
		}

		entriesResponse = append(entriesResponse, itemDataResponse)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get entries successfully.",
		Data:    entriesResponse,
	})
}

// RemoveCronLeave : Remove cron leave
func (ctr *LvController) RemoveCronLeave(c echo.Context) error {
	removeCronLeaveParam := new(param.RemoveCronLeaveParam)

	if err := c.Bind(removeCronLeaveParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(removeCronLeaveParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	ctr.RemoveCron(removeCronLeaveParam.ID)

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove cron successfully.",
	})
}

func (ctr *LvController) ImportBonuses(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	rows, errEx := ex.ReadExcelFile(file)
	if errEx != "" {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: errEx,
		})
	}

	if len(rows) < 2 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "File is empty.",
		})
	}

	var leaveBonusParams []param.LeaveBonus
	userProfile := c.Get("user_profile").(m.User)
	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) > 5 {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid Params",
				Data:    i + 1,
			})
		}

		record := [5]string{"", "", "", "", ""}
		for i := 0; i < len(row); i++ {
			record[i] = row[i]
		}

		for _, elm := range record {
			if elm == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "Invalid field value",
					Data:    i + 1,
				})
			}
		}

		userId, err := ctr.UserRepo.SelectUserIdByEmployeeId(userProfile.OrganizationID, record[0])
		if err != nil {
			if err.Error() == pg.ErrNoRows.Error() {
				return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "User is not exist",
					Data:    i + 1,
				})
			}

			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		leaveBonusParam := new(param.LeaveBonus)
		leaveBonusParam.LeaveBonusTypeID, err = strconv.Atoi(record[1])
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		leaveBonusParam.YearBelong, err = strconv.Atoi(record[2])
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		leaveBonusParam.Hour, err = strconv.ParseFloat(record[4], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		leaveBonusParam.HourRemaining, err = strconv.ParseFloat(record[4], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		leaveBonusParam.UserID = userId
		leaveBonusParam.Reason = record[3]

		leaveBonusParams = append(leaveBonusParams, *leaveBonusParam)
	}

	if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
		userProfile.OrganizationID,
		userProfile.UserProfile.UserID,
		&leaveBonusParams,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Import successfully.",
	})
}

func (ctr *LvController) DownloadTemplate(c echo.Context) error {
	downloadTemplateParam := new(param.DownloadTemplateParam)
	if err := c.Bind(downloadTemplateParam); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(downloadTemplateParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if downloadTemplateParam.TypeFile == "xlsx" {
		if userProfile.LanguageId == cf.EnLanguageId {
			return c.Attachment("internal/platform/excel/template/import-leave.xlsx", "import-leave.xlsx")
		} else if userProfile.LanguageId == cf.VnLanguageId {
			return c.Attachment("internal/platform/excel/template/import-leave-vn.xlsx", "import-leave-vn.xlsx")
		} else {
			return c.Attachment("internal/platform/excel/template/import-leave-jp.xlsx", "import-leave-jp.xlsx")
		}
	}

	if userProfile.LanguageId == cf.EnLanguageId {
		return c.Attachment("internal/platform/excel/template/import-leave.csv", "import-leave.csv")
	} else if userProfile.LanguageId == cf.VnLanguageId {
		return c.Attachment("internal/platform/excel/template/import-leave-vn.csv", "import-leave-vn.csv")
	} else {
		return c.Attachment("internal/platform/excel/template/import-leave-jp.csv", "import-leave-jp.csv")
	}
}

func (ctr *LvController) ExportExcel(c echo.Context) error {
	leaveRequestTypeId, _ := strconv.Atoi(c.FormValue("leave_request_type_id"))
	branch, _ := strconv.Atoi(c.FormValue("branch"))
	exportExcelParams := &param.LeaveRequestListParams{
		UserName:           c.FormValue("username"),
		LeaveRequestTypeID: leaveRequestTypeId,
		Branch:             branch,
		DatetimeLeaveFrom:  c.FormValue("datetime_leave_from"),
		DatetimeLeaveTo:    c.FormValue("datetime_leave_to"),
	}

	userProfile := c.Get("user_profile").(m.User)
	columns := []string{"ulr.user_id", "ulr.leave_request_type_id", "ulr.datetime_leave_from", "ulr.datetime_leave_to", "ulr.subtract_day_off_type_id"}
	leaveRequestRecords, _, err := ctr.LeaveRepo.LeaveRequests(
		userProfile.OrganizationID,
		"ulr.user_id ASC, ulr.datetime_leave_from ASC",
		exportExcelParams,
		false,
		columns...,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "A", 15)
	_ = f.SetColWidth("Sheet1", "B", "C", 30)
	_ = f.SetColWidth("Sheet1", "D", "G", 15)

	titleStyle, _ := f.NewStyle(`{
		"font":{"bold":true, "size":16},
		"alignment":{"horizontal":"center", "vertical":"center"}
	}`)

	contentStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center", "vertical":"center"}}`)

	var categories map[string]string
	if userProfile.LanguageId == cf.EnLanguageId {
		categories = cf.LeaveCategoriesEn
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categories = cf.LeaveCategoriesVn
	} else {
		categories = cf.LeaveCategoriesJp
	}

	leaveCategories := map[string]string{
		"A1": categories["Employee Id"],
		"B1": categories["Full Name"],
		"C1": categories["Leave Request Type"],
		"D1": categories["Date From"],
		"E1": categories["Date To"],
		"F1": categories["Time"],
		"G1": categories["Note"],
	}
	for k, v := range leaveCategories {
		_ = f.SetCellValue("Sheet1", k, v)
	}
	_ = f.SetCellStyle("Sheet1", "A1", "G1", titleStyle)
	_ = f.SetColStyle("Sheet1", "A", contentStyle)
	_ = f.SetColStyle("Sheet1", "D", contentStyle)
	_ = f.SetColStyle("Sheet1", "E", contentStyle)
	_ = f.SetColStyle("Sheet1", "F", contentStyle)
	_ = f.SetColStyle("Sheet1", "G", contentStyle)

	userProfileRecords, err := ctr.UserRepo.SelectEmployeeIdByOrganizationId(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	idx := 2
	var userProfileRecord param.EmployeeIdAndFullName
	for i, record := range leaveRequestRecords {
		pos := idx + i
		for _, upr := range userProfileRecords {
			if record.UserID == upr.UserId {
				userProfileRecord = upr
			}
		}

		var datetimeLeaveFrom, datetimeLeaveTo, hourFrom, minuteFrom, hourTo, minuteTo, leaveTime string
		datetimeLeaveFrom = record.DatetimeLeaveFrom.Format(cf.FormatDateDisplay)
		datetimeLeaveTo = record.DatetimeLeaveTo.Format(cf.FormatDateDisplay)

		if record.LeaveRequestTypeID == cf.LateForWork || record.LeaveRequestTypeID == cf.LeaveEarly || record.LeaveRequestTypeID == cf.GoOutside {
			hourFrom = utils.ConvertTwoChar(record.HourFrom)
			minuteFrom = utils.ConvertTwoChar(record.MinuteFrom)
			hourTo = utils.ConvertTwoChar(record.HourTo)
			minuteTo = utils.ConvertTwoChar(record.MinuteTo)

			if hourFrom != "" && minuteFrom != "" && hourTo != "" && minuteTo != "" {
				leaveTime = hourFrom + ":" + minuteFrom + " - " + hourTo + ":" + minuteTo
			}
		}

		values := map[string]interface{}{
			"A" + strconv.Itoa(pos): userProfileRecord.EmployeeId,
			"B" + strconv.Itoa(pos): userProfileRecord.FullName,
			"C" + strconv.Itoa(pos): cf.LeaveRequestTypes[record.LeaveRequestTypeID],
			"D" + strconv.Itoa(pos): datetimeLeaveFrom,
			"E" + strconv.Itoa(pos): datetimeLeaveTo,
			"F" + strconv.Itoa(pos): leaveTime,
		}

		if record.SubtractDayOffTypeID != cf.Subtract {
			values["G"+strconv.Itoa(pos)] = cf.SubtractDayOffTypes[record.SubtractDayOffTypeID]
		}

		for k, v := range values {
			_ = f.SetCellValue("Sheet1", k, v)
		}
	}

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}

func (ctr *LvController) GetLeaveBonuses(c echo.Context) error {
	params := new(param.GetLeaveBonusParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
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

	records, totalRow, err := ctr.LeaveRepo.SelectLeaveBonuses(userProfile.OrganizationID, params)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var leaveBonusesResponse []map[string]interface{}
	pagination := map[string]interface{}{
		"total_row": totalRow,
	}

	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	for _, record := range records {
		leaveBonus := map[string]interface{}{
			"id":               record.Id,
			"user_id":          record.UserId,
			"created_by":       record.CreatedBy,
			"reason":           record.Reason,
			"year":             record.Year,
			"hour":             record.Hour,
			"leave_bonus_type": cf.LeaveBonusTypes[record.LeaveBonusTypeId],
			"created_at":       record.CreatedAt.In(location).Format(cf.FormatDisplayTimekeeping2),
		}

		leaveBonusesResponse = append(leaveBonusesResponse, leaveBonus)
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
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
	for i := 0; i < len(userRecords); i++ {
		userList[userRecords[i].UserID] = userRecords[i].FullName
	}

	dataResponse := map[string]interface{}{
		"user_leave_bonuses": leaveBonusesResponse,
		"pagination":         pagination,
		"user_list":          userList,
		"leave_bonus_types":  cf.LeaveBonusTypes,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get leave bonuses successful",
		Data:    dataResponse,
	})
}

func (ctr *LvController) GetLeaveBonus(c echo.Context) error {
	params := new(param.GetLeaveBonusByIdParam)
	if err := c.Bind(params); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	leaveBonus, err := ctr.LeaveRepo.SelectLeaveBonusById(params.Id)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	dataResponse := map[string]interface{}{
		"leave_bonus_type_id": leaveBonus.LeaveBonusTypeID,
		"hour":                leaveBonus.Hour,
		"reason":              leaveBonus.Reason,
		"year_belong":         leaveBonus.YearBelong,
		"leave_bonus_types":   cf.LeaveBonusTypes,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get leave bonus successful",
		Data:    dataResponse,
	})
}

func (ctr *LvController) EditLeaveBonuses(c echo.Context) error {
	params := new(param.EditLeaveBonusParam)
	if err := c.Bind(params); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if err := ctr.LeaveRepo.UpdateLeaveBonus(userProfile.UserProfile.UserID, params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update leave bonus successful",
	})
}

func (ctr *LvController) RemoveLeaveBonus(c echo.Context) error {
	params := new(param.RemoveLeaveBonusParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if err := ctr.LeaveRepo.UpdateDeleted(params.Id, params.IsDeleted); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var message string
	if params.IsDeleted {
		message = "Remove leave bonuses successful"
	} else {
		message = "Restore leave bonuses successful"
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: message,
	})
}

func (ctr *LvController) getHolidaysOfYear(organizationId int, year int) []time.Time {
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

func monthsCountSince(createdAtTime time.Time) int {
	now := time.Now()
	currentYear, _, _ := now.Date()
	currentLocation := now.Location()
	lastYear := time.Date(currentYear, 12, 31, 23, 0, 0, 0, currentLocation)
	months := 0
	month := createdAtTime.Month()
	for createdAtTime.Before(lastYear) {
		createdAtTime = createdAtTime.Add(time.Hour * 24)
		nextMonth := createdAtTime.Month()
		if nextMonth != month {
			months++
		}
		month = nextMonth
	}

	return months
}

func (ctr *LvController) UpdateHourRemainingLeave(c echo.Context) error {
	params := new(param.UpdateHourRemainingLeaveParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	validDateRecords, err := ctr.LeaveRepo.SelectValidDateLeave(params.OrgID, params.UserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	hourLeaveTemp := params.HourLeave
	if len(validDateRecords) > 0 {
		for _, validDate := range validDateRecords {
			if hourLeaveTemp > validDate.HourRemaining {
				ctr.LeaveRepo.UpdateHourRemainingLeaveByID(validDate.Id, 0)
			} else {
				ctr.LeaveRepo.UpdateHourRemainingLeaveByID(validDate.Id, validDate.HourRemaining-hourLeaveTemp)
			}
			hourLeaveTemp = hourLeaveTemp - validDate.HourRemaining
			if hourLeaveTemp <= 0 {
				break
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update expire date leave successfully",
	})
}

func (ctr *LvController) AddExpireLeave(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	records, err := ctr.LeaveRepo.SelectStartLeaveCurrentYear(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if len(records) > 0 {
		t := time.Now()
		firstday := time.Date(t.Year()+1, time.Month(userProfile.Organization.ExpirationResetDayOff+1), 1, 0, 0, 0, 0, time.Local)

		for _, record := range records {
			if record.LeaveBonusTypeId == 1 || record.LeaveBonusTypeId == 2 {
				if err := ctr.LeaveRepo.UpdateExpireLeaveDate(record.Id, record.Hour, firstday.Format("2006-01-02 00:04:05")); err != nil {
					return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: "System Error",
					})
				}
			} else if record.LeaveBonusTypeId == 8 {
				if err := ctr.LeaveRepo.UpdateExpireLeaveDate(record.Id, record.Hour, record.UpdatedAt.AddDate(0, 3, 0).Format("2006-01-02 15:04:05")); err != nil {
					return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: "System Error",
					})
				}
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Add expire date successfully",
	})
}
