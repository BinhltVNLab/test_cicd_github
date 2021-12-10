package notification

import (
	"net/http"
	"strconv"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	cr "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cron"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type Controller struct {
	cm.BaseController
	afb.FirebaseCloudMessage
	email.SMTPGoMail
	cr.EtrCron

	NotificationRepo rp.NotificationRepository
	FcmTokenRepo     rp.FcmTokenRepository
	UserRepo         rp.UserRepository
	Cloud            gc.StorageUtility
	OrgRepo          rp.OrgRepository
}

func NewNotificationController(logger echo.Logger, notificationRepo rp.NotificationRepository,
	fcmTokenRepo rp.FcmTokenRepository, userRepo rp.UserRepository, cloud gc.StorageUtility, orgRepo rp.OrgRepository) (ctr *Controller) {
	ctr = &Controller{
		cm.BaseController{},
		afb.FirebaseCloudMessage{},
		email.SMTPGoMail{},
		cr.EtrCron{},
		notificationRepo,
		fcmTokenRepo,
		userRepo,
		cloud,
		orgRepo,
	}
	ctr.Init(logger)
	ctr.InitCron("Asia/Ho_Chi_Minh")
	ctr.InitFcm()
	return
}

func (ctr *Controller) EditNotificationStatusRead(c echo.Context) error {
	params := new(param.UpdateNotificationStatusReadParam)
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

	userProfile := c.Get("user_profile").(m.User)
	if params.Receiver != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to edit notification status",
		})
	}

	if err := ctr.NotificationRepo.UpdateNotificationStatusRead(userProfile.OrganizationID, params.Receiver); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit notification status to read successful",
	})
}

func (ctr *Controller) EditNotificationStatus(c echo.Context) error {
	params := new(param.UpdateNotificationStatusParam)
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

	userProfile := c.Get("user_profile").(m.User)
	if params.Receiver != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to edit notification status",
		})
	}

	count, err := ctr.NotificationRepo.CheckNotificationExist(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Notification does not exist",
		})
	}

	err = ctr.NotificationRepo.UpdateNotificationStatus(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit notification status successful",
	})
}

func (ctr *Controller) GetNotifications(c echo.Context) error {
	params := new(param.GetNotificationsParam)
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

	userProfile := c.Get("user_profile").(m.User)
	if params.Receiver != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to get notifications",
		})
	}

	if params.RowPerPage == 0 {
		params.CurrentPage = 1
		params.RowPerPage = 8
	}

	records, totalRow, err := ctr.NotificationRepo.SelectNotifications(userProfile.OrganizationID, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": params.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": params.RowPerPage,
	}

	var notifications []map[string]interface{}
	var days []time.Time
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	for _, record := range records {
		days = append(days, record.CreatedAt)
		data := map[string]interface{}{
			"id":           record.Id,
			"sender":       record.Sender,
			"content":      record.Content,
			"status":       record.Status,
			"redirect_url": record.RedirectUrl,
			"created_at":   record.CreatedAt.In(location).Format(cf.FormatDateNoSec),
		}

		var base64Img []byte
		if record.AvatarSender != "" {
			base64Img, err = ctr.Cloud.GetFileByFileName(record.AvatarSender, cf.AvatarFolderGCS)
			if err != nil {
				ctr.Logger.Error(err)
			}
		}
		data["avatar_sender"] = base64Img

		notifications = append(notifications, data)
	}

	var smallestDay time.Time
	if len(days) > 0 {
		smallestDay = days[0]
		for i := 1; i < len(days); i++ {
			if days[i].Sub(smallestDay) < 0 {
				smallestDay = days[i]
			}
		}
	}

	dataResponse := map[string]interface{}{
		"pagination":              pagination,
		"notification_status_map": cf.NotificationStatusMap,
		"notifications":           notifications,
		"smallest_day":            smallestDay.Format(cf.FormatDate),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get notifications successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) GetTotalNotificationsUnread(c echo.Context) error {
	params := new(param.GetTotalNotificationsUnreadParam)
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

	clientTime := calendar.ParseTime(cf.FormatDate, params.ClientTime)
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	serverTime := calendar.ParseTime(cf.FormatDate, time.Now().In(location).Format(cf.FormatDate))
	duration, _ := time.ParseDuration("0h0m3s")
	if serverTime.Sub(clientTime) > duration || serverTime.Sub(clientTime) < -duration {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Server could not precess the request",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.NotificationRepo.CountNotificationsUnRead(userProfile.OrganizationID, userProfile.UserProfile.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get total notifications unread successful",
		Data:    count,
	})
}

func (ctr *Controller) RemoveNotification(c echo.Context) error {
	params := new(param.RemoveNotificationParam)
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

	count, err := ctr.NotificationRepo.CheckNotificationExist(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Notification does not exist",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if params.Receiver != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to remove notification",
		})
	}

	if err := ctr.NotificationRepo.DeleteNotification(params.Id); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove notification successful",
	})
}

func (ctr *Controller) SendNotiRequestById(c echo.Context) error {
	params := new(param.SendNotiRequestParam)
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

	notiRequest, err := ctr.NotificationRepo.FindNotiRequestById(params.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Notification request does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error: " + err.Error(),
		})
	}

	if notiRequest.Status != m.NotiRequestStatusInitial && notiRequest.Status != m.NotiRequestStatusFailedProcessed {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "That notification request was not INITIAL",
		})
	}

	// To avoid concurrency, update status first
	if err := ctr.NotificationRepo.UpdateNotiRequestStatus(notiRequest.ID, notiRequest.Status, m.NotiRequestStatusProcessing); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error: " + err.Error(),
		})
	}
	notiRequest.Status = m.NotiRequestStatusProcessing

	// Handle send notification by FCM
	notifications, err := ctr.NotificationRepo.GetNotificationsByNotiRequest(notiRequest)
	if err != nil {
		_ = ctr.NotificationRepo.UpdateNotiRequestStatus(notiRequest.ID, notiRequest.Status, m.NotiRequestStatusFailedProcessed)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error: " + err.Error(),
		})
	}

	receiverIds := make([]int, 0, len(notifications))
	senderId := 0
	if len(notifications) > 0 {
		senderId = notifications[0].Sender
	}
	for _, _noti := range notifications {
		receiverIds = append(receiverIds, _noti.Receiver)
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(receiverIds, senderId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		_ = ctr.NotificationRepo.UpdateNotiRequestStatus(notiRequest.ID, notiRequest.Status, m.NotiRequestStatusFailedProcessed)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error: " + err.Error(),
		})
	}

	if len(registrationTokens) > 0 {
		var body, link, title string
		if len(notifications) > 0 {
			body = notifications[0].SenderName + " " + notifications[0].Content
			link, title = notifications[0].RedirectUrl, notifications[0].Title
		}
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, title, body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	// Handle send email notification base on email & password of organization
	emailNotiRequests, err := ctr.NotificationRepo.GetEmailNotificationsByNotiRequest(notiRequest)
	if err != nil {
		_ = ctr.NotificationRepo.UpdateNotiRequestStatus(notiRequest.ID, notiRequest.Status, m.NotiRequestStatusFailedProcessed)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error: " + err.Error(),
		})
	}

	for _, _value := range emailNotiRequests {
		if _value.Email != "" && _value.EmailPassword != "" {
			ctr.InitSmtp(_value.Email, _value.EmailPassword)
			emails, _ := ctr.UserRepo.SelectEmailByUserIds(_value.ToUserIds)
			if len(emails) > 0 {
				sampleData := new(param.SampleData)
				sampleData.SendTo = emails
				sampleData.Content = _value.Content
				sampleData.URL = _value.Url
				if err := ctr.SendMail(
					_value.Subject,
					sampleData,
					_value.Template,
				); err != nil {
					ctr.Logger.Error(err)
				}
			}
		}
	}

	_ = ctr.NotificationRepo.UpdateNotiRequestStatus(notiRequest.ID, notiRequest.Status, m.NotiRequestStatusSucceedProcessed)
	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Send notification request successful",
	})
}

// NotiEventRemind : Cron annual leave bonus with run year-01-01 00:00:00 and clear old leave with run year-04-01 00:00:00
func (ctr *Controller) NotiEventRemind(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

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

	var birthdayNotiParams []param.BirthdayNoti
	var companyJoinedParams []param.CompanyJoinedNoti
	var contractRemindParams []param.RemindNoti
	t := time.Now()

	for _, user := range users {
		birthday, _ := time.Parse(cf.FormatDateDatabase, user.Birthday)
		birthdayThisYear := time.Date(t.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)
		threeDaysBeforeBirthday := birthdayThisYear.AddDate(0, 0, -30)
		if t.After(threeDaysBeforeBirthday) && t.Before(birthdayThisYear.AddDate(0, 0, 1)) {
			birthdayNotiParam := new(param.BirthdayNoti)
			birthdayNotiParam.FullName = user.FullName
			birthdayNotiParam.Birthday = user.Birthday
			birthdayNotiParams = append(birthdayNotiParams, *birthdayNotiParam)
		}

		joinedDateThisYear := time.Date(t.Year(), user.CompanyJoinedDate.Month(), user.CompanyJoinedDate.Day(), 0, 0, 0, 0, time.UTC)
		threeDaysBeforeJoinedDate := joinedDateThisYear.AddDate(0, 0, -4)
		if t.After(threeDaysBeforeJoinedDate) && t.Before(joinedDateThisYear.AddDate(0, 0, 1)) {
			companyJoinedParam := new(param.CompanyJoinedNoti)
			companyJoinedParam.FullName = user.FullName
			companyJoinedParam.CompanyJoinedDate = user.CompanyJoinedDate.Format(cf.FormatDateDatabase)
			companyJoinedParams = append(companyJoinedParams, *companyJoinedParam)
		}

		sevenDaysBeforeContractRemind := user.ContractExpirationDate.AddDate(0, 0, -8)
		if t.After(sevenDaysBeforeContractRemind) && t.Before(user.ContractExpirationDate.AddDate(0, 0, 1)) {
			remindNotiParam := new(param.RemindNoti)
			remindNotiParam.FullName = user.FullName
			remindNotiParam.ContractExpirationDate = user.ContractExpirationDate.Format(cf.FormatDateDatabase)
			contractRemindParams = append(contractRemindParams, *remindNotiParam)
		}
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	dataResponse := map[string]interface{}{
		"bithday_list":           birthdayNotiParams,
		"company_join_date_list": companyJoinedParams,
		"contract_remind_list":   contractRemindParams,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create cron event successfully.",
		Data:    dataResponse,
	})
}

// CronLeaveBonus : Cron annual leave bonus with run year-01-01 00:00:00 and clear old leave with run year-04-01 00:00:00
func (ctr *Controller) CronNotiEventRemind(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	user, _ := ctr.UserRepo.GetUser(userProfile.LanguageId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	org, err := ctr.OrgRepo.SelectEmailAndPassword(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	usersIdGm, err := ctr.UserRepo.SelectIdsOfGM(userProfile.OrganizationID)
	emails, _ := ctr.UserRepo.SelectEmailByUserIds(usersIdGm)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	_, err = ctr.AddFuncCron("0 6 * * *", "Notify event and remind cron", func() {
		if org.Email != "" && org.EmailPassword != "" {
			ctr.InitSmtp(org.Email, org.EmailPassword)
		}

		var birthdayNotiParams []param.BirthdayNoti
		var companyJoinedParams []param.CompanyJoinedNoti
		var contractRemindParams []param.RemindNoti
		t := time.Now()

		for _, user := range users {
			birthday, _ := time.Parse(cf.FormatDateDatabase, user.Birthday)
			birthdayThisYear := time.Date(t.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)
			threeDaysBeforeBirthday := birthdayThisYear.AddDate(0, 0, -4)
			if t.After(threeDaysBeforeBirthday) && t.Before(birthdayThisYear.AddDate(0, 0, 1)) {
				birthdayNotiParam := new(param.BirthdayNoti)
				birthdayNotiParam.FullName = user.FullName
				birthdayNotiParam.Birthday = user.Birthday
				birthdayNotiParams = append(birthdayNotiParams, *birthdayNotiParam)
			}

			joinedDateThisYear := time.Date(t.Year(), user.CompanyJoinedDate.Month(), user.CompanyJoinedDate.Day(), 0, 0, 0, 0, time.UTC)
			threeDaysBeforeJoinedDate := joinedDateThisYear.AddDate(0, 0, -4)
			if t.After(threeDaysBeforeJoinedDate) && t.Before(joinedDateThisYear.AddDate(0, 0, 1)) {
				companyJoinedParam := new(param.CompanyJoinedNoti)
				companyJoinedParam.FullName = user.FullName
				companyJoinedParam.CompanyJoinedDate = user.CompanyJoinedDate.Format(cf.FormatDateDatabase)
				companyJoinedParams = append(companyJoinedParams, *companyJoinedParam)
			}

			sevenDaysBeforeContractRemind := user.ContractExpirationDate.AddDate(0, 0, -8)
			if t.After(sevenDaysBeforeContractRemind) && t.Before(user.ContractExpirationDate.AddDate(0, 0, 1)) {
				remindNotiParam := new(param.RemindNoti)
				remindNotiParam.FullName = user.FullName
				remindNotiParam.ContractExpirationDate = user.ContractExpirationDate.Format(cf.FormatDateDatabase)
				contractRemindParams = append(contractRemindParams, *remindNotiParam)
			}
		}
		if (user.LanguageId == 1) {
			if len(birthdayNotiParams) > 0 {
				for _, item := range birthdayNotiParams {
					birthday, _ := time.Parse(cf.FormatDateDatabase, item.Birthday)
					birthdayThisYear := time.Date(t.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := birthdayThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Happy birthday to " + item.FullName + "!",
						"Happy birthday to " + item.FullName + "!",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content = `Have a nice day!
						
						Happy birthday `+item.FullName+`! Everyone at `+userProfile.Organization.Name+` wishes you the best for this special day and for the year ahead.
						
						Congratulations, and thank you for your contributions. We’re honoured to have you with us.
						
						Best regards!`
						if err := ctr.SendMail("[All-members] Happy birthday to "+item.FullName+"!", sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
			if len(companyJoinedParams) > 0 {
				for _, item := range companyJoinedParams {
					joinedDate, _ := time.Parse(cf.FormatDateDatabase, item.CompanyJoinedDate)
					joinDateThisYear := time.Date(t.Year(), joinedDate.Month(), joinedDate.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := joinDateThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content = `Have a nice day!
	
						Today, we're celebrating ` + strconv.Itoa(t.Year() - joinedDate.Year())+`-year anniversary since ` +item.FullName +` joined `+userProfile.Organization.Name+` company. Congratulations and thanks for your contributions.
	
						We hope together, we can make the company grow bigger and be a better workspace for every members.`
	
						if err := ctr.SendMail("[All-members] Celebrating day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
		}
		if (user.LanguageId == 3) {
			if len(birthdayNotiParams) > 0 {
				for _, item := range birthdayNotiParams {
					birthday, _ := time.Parse(cf.FormatDateDatabase, item.Birthday)
					birthdayThisYear := time.Date(t.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := birthdayThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Happy birthday to " + item.FullName + "!",
						"Happy birthday to " + item.FullName + "!",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content = `Chúc một ngày tốt lành!
						
						Chúc Mừng Sinh Nhật `+item.FullName+`! `+userProfile.Organization.Name+`
						Chúc bạn tuổi mới nhiều sức khỏe, hạnh phúc, giàu nhiệt huyết để gặt hái thêm nhiều thành công mới trong công việc và cuộc sống. `
						
						if err := ctr.SendMail("[All-members] Happy birthday to "+item.FullName+"!", sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
			if len(companyJoinedParams) > 0 {
				for _, item := range companyJoinedParams {
					joinedDate, _ := time.Parse(cf.FormatDateDatabase, item.CompanyJoinedDate)
					joinDateThisYear := time.Date(t.Year(), joinedDate.Month(), joinedDate.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := joinDateThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content = `Chúc một ngày tốt lành!
	
						Hôm nay, ngày tròn ` + strconv.Itoa(t.Year() - joinedDate.Year())+`năm, kỷ niệm ngày ` +item.FullName +` gia nhập công ty `+userProfile.Organization.Name+` xin chúc mừng và cảm ơn những đóng góp của bạn trong thời gian vừa qua.
	
						Hy vọng chúng ta có thể cùng công ty ngày càng phát triển và trở thành không gian làm việc tốt hơn cho mọi thành viên.`
	
						if err := ctr.SendMail("[All-members] Celebrating day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
		}
		if (user.LanguageId == 2) {
			if len(birthdayNotiParams) > 0 {
				for _, item := range birthdayNotiParams {
					birthday, _ := time.Parse(cf.FormatDateDatabase, item.Birthday)
					birthdayThisYear := time.Date(t.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := birthdayThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(birthdayThisYear) {
						eventDateStr = birthdayThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Happy birthday to " + item.FullName + "!",
						"Happy birthday to " + item.FullName + "!",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content = `お誕生日おめでとうございます。 この1年が素晴らしい年でありますように `
						 
						if err := ctr.SendMail("[All-members] Happy birthday to "+item.FullName+"!", sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
			if len(companyJoinedParams) > 0 {
				for _, item := range companyJoinedParams {
					joinedDate, _ := time.Parse(cf.FormatDateDatabase, item.CompanyJoinedDate)
					joinDateThisYear := time.Date(t.Year(), joinedDate.Month(), joinedDate.Day(), 0, 0, 0, 0, time.UTC)
					eventDateStr := joinDateThisYear.Format(cf.FormatDateDatabase)
	
					if calendar.IsSaturday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -1).Format(cf.FormatDateDatabase)
					}
	
					if calendar.IsSunday(joinDateThisYear) {
						eventDateStr = joinDateThisYear.AddDate(0, 0, -2).Format(cf.FormatDateDatabase)
					}
	
					eventDateParse, _ := time.Parse(cf.FormatDateDatabase, eventDateStr)
	
					if t.Month() == eventDateParse.Month() && t.Day() == eventDateParse.Day() {
						_ = calendar.AddEventRemind(
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						"Celebrating the day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						eventDateStr,
						eventDateStr)
	
						sampleData := new(param.SampleData)
						sampleData.SendTo = []string{cf.NoticeEmail}
						sampleData.Content =  strconv.Itoa(t.Year() - joinedDate.Year()) +`入社記念日おめでとうございます！今年もよろしくお願いします。`
	
						if err := ctr.SendMail("[All-members] Celebrating day "+ item.FullName + " joined "+ userProfile.Organization.Name  +" company ",
						sampleData, cf.LeaveRequestTemplate); err != nil {
							ctr.Logger.Error(err)
						}
					}
				}
			}
		}		

		if len(contractRemindParams) > 0 {
			for _, item := range contractRemindParams {
				contractExpirationDate, _ := time.Parse(cf.FormatDateDatabase, item.ContractExpirationDate)
				sevenDaysBeforeContractRemind := contractExpirationDate.AddDate(0, 0, -8)

				if t.After(sevenDaysBeforeContractRemind) && t.Before(contractExpirationDate.AddDate(0, 0, 1)) {
					sampleData := new(param.SampleData)
					sampleData.SendTo = emails
					sampleData.Content = item.FullName + "'s contract will expire "+ contractExpirationDate.Format(cf.FormatDateDisplay) +
					". Please renew a new contract for " + item.FullName
					if err := ctr.SendMail("[Notification] Contract extension for " +item.FullName, sampleData, cf.LeaveRequestTemplate); err != nil {
						ctr.Logger.Error(err)
					}
				}
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
		Message: "Create cron notify event, remind successfully.",
	})
}

func (ctr *Controller) CronNotifyEventRemindStart(c echo.Context) error {
	ctr.StartCron()
	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Start cron notify event, remind successfully.",
	})
}