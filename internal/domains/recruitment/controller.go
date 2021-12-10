package recruitment

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type Controller struct {
	cm.BaseController
	email.SMTPGoMail
	afb.FirebaseCloudMessage

	Cloud            gc.StorageUtility
	RecruitmentRepo  rp.RecruitmentRepository
	ProjectRepo      rp.ProjectRepository
	BranchRepo       rp.BranchRepository
	UserRepo         rp.UserRepository
	NotificationRepo rp.NotificationRepository
	FcmTokenRepo     rp.FcmTokenRepository
}

func NewRecruitmentController(
	logger echo.Logger,
	cloud gc.StorageUtility,
	recruitmentRepo rp.RecruitmentRepository,
	projectRepo rp.ProjectRepository,
	branchRepo rp.BranchRepository,
	userRepo rp.UserRepository,
	notificationRepo rp.NotificationRepository,
	fcmTokenRepo rp.FcmTokenRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, email.SMTPGoMail{}, afb.FirebaseCloudMessage{}, cloud,
		recruitmentRepo, projectRepo, branchRepo, userRepo, notificationRepo, fcmTokenRepo}
	ctr.Init(logger)
	ctr.InitFcm()
	return
}

func (ctr *Controller) CreateJob(c echo.Context) error {
	params := new(param.CreateJobParam)
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

	startDate := calendar.ParseTime(cf.FormatDateDatabase, params.StartDate)
	expiryDate := calendar.ParseTime(cf.FormatDateDatabase, params.ExpiryDate)
	if startDate.Sub(expiryDate) > 0 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Expiry date must be greater start date",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	body, link, err := ctr.RecruitmentRepo.InsertJob(userProfile.OrganizationID, userProfile.UserProfile.UserID, params, ctr.NotificationRepo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(params.Assignees, userProfile.UserProfile.UserID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
	if len(registrationTokens) > 0 {
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	if userProfile.Organization.Email != "" && userProfile.Organization.EmailPassword != "" {
		ctr.InitSmtp(userProfile.Organization.Email, userProfile.Organization.EmailPassword)
		emails, _ := ctr.UserRepo.SelectEmailByUserIds(params.Assignees)
		sampleData := new(param.SampleData)
		sampleData.SendTo = emails
		sampleData.Content = "Hi there, you have been assigned to a recruiting job. Please click the button below for more information"
		sampleData.URL = os.Getenv("BASE_SPA_URL") + link
		if err := ctr.SendMail(
			"【Notification】【Micro erp】Assigned recruitment",
			sampleData,
			cf.Recruitment,
		); err != nil {
			ctr.Logger.Error(err)
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create job successful",
	})
}

func (ctr *Controller) CreateDetailJob(c echo.Context) error {
	params := new(param.CreateDetailJobParam)
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

	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, []string{"organization_id", "assignees"}...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCUD_DetailedJob(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to do this action",
		})
	}

	detailedRecruitment := m.DetailedJobRecruitment{
		RecruitmentID:     params.RecruitmentId,
		Amount:            params.Amount,
		Address:           params.Address,
		Place:             params.Place,
		Role:              params.Role,
		Gender:            params.Gender,
		TypeOfWork:        fmt.Sprintf("%v", params.TypeOfWork), // TODO re-check here
		Experience:        params.Experience,
		SalaryType:        params.SalaryType,
		SalaryFrom:        params.SalaryFrom,
		SalaryTo:          params.SalaryTo,
		ProfileRecipients: params.ProfileRecipients,
		Email:             params.Email,
		PhoneNumber:       params.PhoneNumber,
		Description:       params.Description,
	}

	if err := ctr.RecruitmentRepo.InsertDetailedJobRecruitmentFromModel(&detailedRecruitment); err != nil {
		ctr.Logger.Errorf("Insert detailed job got error: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when insert new detailed job recruitment",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create detail job successful",
	})
}

func (ctr *Controller) UpdateDetailJob(c echo.Context) error {
	params := new(param.UpdateDetailJobParam)
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

	detailJob, err := ctr.RecruitmentRepo.FindDetailedJobRecruitmentById(params.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Detailed Job Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when query detail job recruitment",
		})
	}

	recruitment, err := ctr.RecruitmentRepo.SelectJob(detailJob.RecruitmentID, []string{"organization_id", "assignees"}...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCUD_DetailedJob(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to do this action",
		})
	}

	updateDetailedRecruitment := m.DetailedJobRecruitment{
		BaseModel: cm.BaseModel{
			ID: detailJob.ID,
		},
		RecruitmentID:     detailJob.RecruitmentID,
		Amount:            params.Amount,
		Address:           params.Address,
		Place:             params.Place,
		Role:              params.Role,
		Gender:            params.Gender,
		TypeOfWork:        fmt.Sprintf("%v", params.TypeOfWork), // TODO re-check here
		Experience:        params.Experience,
		SalaryType:        params.SalaryType,
		SalaryFrom:        params.SalaryFrom,
		SalaryTo:          params.SalaryTo,
		ProfileRecipients: params.ProfileRecipients,
		Email:             params.Email,
		PhoneNumber:       params.PhoneNumber,
		Description:       params.Description,
	}

	if err := ctr.RecruitmentRepo.UpdateDetailedJobRecruitmentFromModel(&updateDetailedRecruitment); err != nil {
		ctr.Logger.Errorf("Update detailed job got error: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when update detailed job recruitment",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update detail job successful",
	})
}

func (ctr *Controller) EditJob(c echo.Context) error {
	params := new(param.EditJobParam)
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

	if params.StartDate != "" && params.ExpiryDate != "" {
		startDate := calendar.ParseTime(cf.FormatDateDatabase, params.StartDate)
		expiryDate := calendar.ParseTime(cf.FormatDateDatabase, params.ExpiryDate)
		if startDate.Sub(expiryDate) > 0 {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Expiry date must be greater start date",
			})
		}
	}

	count, err := ctr.RecruitmentRepo.CountJob(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Job does not exist",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if err := ctr.RecruitmentRepo.UpdateJob(userProfile.OrganizationID, params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update job successful",
	})
}

func (ctr *Controller) GetJobs(c echo.Context) error {
	params := new(param.GetJobsParam)
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
	records, totalRow, err := ctr.RecruitmentRepo.SelectJobs(userProfile.OrganizationID, params)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
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

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	if len(branchRecords) > 0 {
		for _, record := range branchRecords {
			branches[record.Id] = record.Name
		}
	}

	if params.BranchId != 0 {
		if _, ok := branches[params.BranchId]; !ok {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Branch does not exist",
			})
		}
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var usersId []int
	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			usersId = append(usersId, user.UserID)
			users[user.UserID] = user.FullName
		}
	}

	var recruitments []map[string]interface{}
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	for _, record := range records {
		data := map[string]interface{}{
			"id":          record.Id,
			"job_name":    record.JobName,
			"start_date":  record.StartDate.In(location).Format(cf.FormatDateDatabase),
			"expiry_date": record.ExpiryDate.In(location).Format(cf.FormatDateDatabase),
			"branch_ids":  record.BranchIds,
			"assignees":   record.Assignees,
		}

		recruitments = append(recruitments, data)
	}

	dataResponse := map[string]interface{}{
		"recruitments": recruitments,
		"branches":     branches,
		"pagination":   pagination,
		"users":        users,
		"avatars":      ctr.SelectAssigneeAndAvatars(usersId),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get jobs successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) GetJob(c echo.Context) error {
	params := new(param.GetJobParam)
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

	columns := []string{"job_name", "start_date", "expiry_date", "branch_ids", "assignees", "organization_id", "detail_job_file_name"}
	rcRecord, err := ctr.RecruitmentRepo.SelectJob(params.Id, columns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Job does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID &&
		userProfile.RoleID != cf.ManagerRoleID &&
		!utils.FindIntInSlice(rcRecord.Assignees, userProfile.UserProfile.UserID) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to view job",
		})
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	if len(branchRecords) > 0 {
		for _, record := range branchRecords {
			branches[record.Id] = record.Name
		}
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var usersId []int
	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			usersId = append(usersId, user.UserID)
			users[user.UserID] = user.FullName
		}
	}

	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var filePath string
	if rcRecord.DetailJobFileName != "" {
		filePath = "https://storage.googleapis.com/" + os.Getenv("GOOGLE_STORAGE_BUCKET") + "/" + configs.CVFOLDERGCS +
			strconv.Itoa(rcRecord.OrganizationId) + "/" +
			strings.Replace(rcRecord.JobName, " ", "_", -1) + rcRecord.DetailJobFileName
	}

	recruitment := map[string]interface{}{
		"job_name":             rcRecord.JobName,
		"start_date":           rcRecord.StartDate.Format(cf.FormatDateDatabase),
		"expiry_date":          rcRecord.ExpiryDate.Format(cf.FormatDateDatabase),
		"branch_ids":           rcRecord.BranchIds,
		"assignees":            rcRecord.Assignees,
		"detail_job_file_name": rcRecord.DetailJobFileName,
		"detail_job_file_path": filePath,
	}

	dataResponse := map[string]interface{}{
		"recruitment": recruitment,
		"branches":    branches,
		"users":       users,
		"avatars":     ctr.SelectAssigneeAndAvatars(usersId),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get job successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) RemoveJob(c echo.Context) error {
	params := new(param.RemoveJobParam)
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

	count, err := ctr.RecruitmentRepo.CountJob(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Job does not exist",
		})
	}

	if err := ctr.RecruitmentRepo.DeleteJob(params.Id); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove job successful",
	})
}

func (ctr *Controller) UploadCv(c echo.Context) error {
	params := new(param.UploadCvParam)
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

	columns := []string{"job_name"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, columns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if len(params.CvFields) > 0 {
		for _, cv := range params.CvFields {
			err := ctr.Cloud.UploadFileToCloud(
				cv.Content,
				cv.FileName,
				cf.CVFOLDERGCS+strconv.Itoa(userProfile.OrganizationID)+"/"+
					strings.Replace(recruitment.JobName, " ", "_", -1),
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System Error",
					Data:    err,
				})
			}
		}
	}

	body, link, err := ctr.RecruitmentRepo.InsertCv(userProfile.OrganizationID, userProfile.UserProfile.UserID, params, ctr.NotificationRepo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(params.Assignees, userProfile.UserProfile.UserID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
	if len(registrationTokens) > 0 {
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	if userProfile.Organization.Email != "" && userProfile.Organization.EmailPassword != "" {
		ctr.InitSmtp(userProfile.Organization.Email, userProfile.Organization.EmailPassword)
		emails, _ := ctr.UserRepo.SelectEmailByUserIds(params.Assignees)
		sampleData := new(param.SampleData)
		sampleData.SendTo = emails
		sampleData.Content = "Hi there, " + body + ". Please click the button below for more information"
		sampleData.URL = os.Getenv("BASE_SPA_URL") + link
		if err := ctr.SendMail(
			"【Notification】【Micro erp】Update cv status",
			sampleData,
			cf.Recruitment,
		); err != nil {
			ctr.Logger.Error(err)
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Upload cv successful",
	})
}

func (ctr *Controller) CreateCv(c echo.Context) error {
	params := new(param.CreateCvParam)
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

	t, err := time.Parse(cf.FormatDateDatabase, params.DateReceiptCv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid value for field date_receipt_cv",
		})
	}
	columns := []string{"job_name", "organization_id", "assignees"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, columns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCRUD_CV(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to do this action",
		})
	}

	directoryCloud := recruitment.GCSCVDirectory()
	millisecondTimeNow := int(time.Now().UTC().UnixNano() / int64(time.Millisecond))
	params.FileName = strings.Replace(strconv.Itoa(millisecondTimeNow)+"_"+params.FileName,
		" ", "_", -1)

	if err := ctr.Cloud.UploadFileToCloud(
		params.FileContent,
		params.FileName,
		directoryCloud,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	cv := m.Cv{
		RecruitmentId:   params.RecruitmentId,
		MediaId:         params.MediaId,
		FileName:        params.FileName,
		FullName:        params.FullName,
		PhoneNumber:     params.PhoneNumber,
		Email:           params.Email,
		DateReceiptCv:   t,
		InterviewMethod: params.InterviewMethod,
		Salary:          params.Salary,
		ContactLink:     params.ContactLink,
		MediaIdOther:    params.MediaIdOther,
	}
	
	body, link, err := ctr.RecruitmentRepo.CreateCvNoti(&cv, userProfile.OrganizationID, userProfile.UserProfile.UserID, params, ctr.NotificationRepo)
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(params.Assignees, userProfile.UserProfile.UserID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
	if len(registrationTokens) > 0 {
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	if userProfile.Organization.Email != "" && userProfile.Organization.EmailPassword != "" {
		ctr.InitSmtp(userProfile.Organization.Email, userProfile.Organization.EmailPassword)
		emails, _ := ctr.UserRepo.SelectEmailByUserIds(params.Assignees)
		sampleData := new(param.SampleData)
		sampleData.SendTo = emails
		sampleData.Content = "Hi there, " + body + ". Please click the button below for more information"
		sampleData.URL = os.Getenv("BASE_SPA_URL") + link
		if err := ctr.SendMail(
			"【Notification】【Micro erp】Create new CV",
			sampleData,
			cf.Recruitment,
		); err != nil {
			ctr.Logger.Error(err)
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create cv successful",
	})
}

func (ctr *Controller) GetCvs(c echo.Context) error {
	params := new(param.GetCvsParam)
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

	rcColumns := []string{"assignees", "job_name"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, rcColumns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID &&
		userProfile.RoleID != cf.ManagerRoleID &&
		!utils.FindIntInSlice(recruitment.Assignees, userProfile.UserProfile.UserID) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to get cvs",
		})
	}

	cvsRecord, totalRow, err := ctr.RecruitmentRepo.SelectCvs(params.RecruitmentId, params)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
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

	var responses []map[string]interface{}
	for _, record := range cvsRecord {

		base64ContentFile, err := ctr.Cloud.GetFileByFileName(record.FileName,
			cf.CVFOLDERGCS+strconv.Itoa(userProfile.OrganizationID)+"/")

		if err != nil {
			ctr.Logger.Error(err)
			base64ContentFile = nil
		}
		fmt.Println("responses")
		res := map[string]interface{}{
			"id":              record.Id,
			"full_name":       record.FullName,
			"date_receipt_cv": record.DateReceiptCv.Format(cf.FormatDateDisplay),
			"media_id":        record.MediaID,
			"updated_at":      record.LastUpdatedAt.Format(cf.FormatDateDisplay),
			"status":          record.StatusCV,
			"file_content":    base64ContentFile,
			"media_id_other":	record.MediaIDOther,
		}
		responses = append(responses, res)
	}

	dataResponse := map[string]interface{}{
		"pagination": pagination,
		"cvs":        responses,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get cvs successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) UpdateCv(c echo.Context) error {
	params := new(param.UpdateCvParam)
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

	t, err := time.Parse(cf.FormatDateDatabase, params.DateReceiptCv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid value for field date_receipt_cv",
		})
	}

	cv, err := ctr.RecruitmentRepo.FindCvById(params.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Cv does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when FindCvById",
		})
	}

	var (
		oldRecruitment, newRecruitment m.Recruitment
		columns                        = []string{"job_name", "organization_id", "assignees"}
	)

	oldRecruitment, err = ctr.RecruitmentRepo.SelectJob(cv.RecruitmentId, columns...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when query old Recruitment",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !oldRecruitment.IsAllowCRUD_CV(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to do this action",
		})
	}

	if params.RecruitmentId == cv.RecruitmentId {
		newRecruitment = oldRecruitment
	} else {
		newRecruitment, err = ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, columns...)
		if err != nil {
			if err.Error() == pg.ErrNoRows.Error() {
				return c.JSON(http.StatusNotFound, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "Recruitment does not exist",
				})
			}

			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error when query Recruitment",
			})
		}
	}

	if params.FileContent != "" {
		// Upload new CV file to GCS first
		directoryCloud := cf.CVFOLDERGCS + strconv.Itoa(userProfile.OrganizationID) + "/" +
			strings.Replace(newRecruitment.JobName, " ", "_", -1) // upload file to dir base on newRecruitment
		millisecondTimeNow := int(time.Now().UTC().UnixNano() / int64(time.Millisecond))
		params.FileName = strings.Replace(strconv.Itoa(millisecondTimeNow)+"_"+params.FileName,
			" ", "_", -1)

		if err := ctr.Cloud.UploadFileToCloud(
			params.FileContent,
			params.FileName,
			directoryCloud,
		); err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	} else {
		params.FileName = cv.FileName
	}

	updateCv := m.Cv{
		BaseModel: cm.BaseModel{
			ID: cv.ID,
		},
		RecruitmentId:   params.RecruitmentId,
		MediaId:         params.MediaId,
		MediaIdOther: params.MediaIdOther,
		FileName:        params.FileName,
		FullName:        params.FullName,
		PhoneNumber:     params.PhoneNumber,
		Email:           params.Email,
		Salary:          params.Salary,
		DateReceiptCv:   t,
		InterviewMethod: params.InterviewMethod,
		ContactLink:     params.ContactLink,
	}

	if err := ctr.RecruitmentRepo.UpdateCv(&updateCv); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System got error when update CV",
		})
	} else {
		// Everything of main flow is success, now to delete old file on GCS if have update
		if params.FileContent != "" {
			oldDirectory := cf.CVFOLDERGCS + strconv.Itoa(userProfile.OrganizationID) + "/" +
				strings.Replace(oldRecruitment.JobName, " ", "_", -1)

			if err := ctr.Cloud.DeleteFileCloud(cv.FileName, oldDirectory); err != nil {
				ctr.Logger.Error(err)
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update cv successful",
	})
}

func (ctr *Controller) GetCvById(c echo.Context) error {
	params := new(param.GetCVByIdParam)
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

	cv, err := ctr.RecruitmentRepo.FindCvById(params.Id)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Cv does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	recruitment, err := ctr.RecruitmentRepo.SelectJob(cv.RecruitmentId, []string{"organization_id", "job_name", "assignees"}...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCRUD_CV(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to view detail of this CV",
		})
	}

	byteArr, err := ctr.Cloud.GetFileByFileName(cv.FileName, recruitment.GCSCVDirectory())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when download CV from GCS",
		})
	}

	var filePath string
	if cv.FileName != "" {
		filePath = "https://storage.googleapis.com/" + os.Getenv("GOOGLE_STORAGE_BUCKET") + "/" + configs.CVFOLDERGCS +
			strconv.Itoa(userProfile.OrganizationID) + "/" +
			strings.Replace(recruitment.JobName, " ", "_", -1) + cv.FileName
	}

	resp := map[string]interface{}{
		"recruitment_id":   cv.RecruitmentId,
		"full_name":        cv.FullName,
		"phone_number":     cv.PhoneNumber,
		"email":            cv.Email,
		"salary":           cv.Salary,
		"date_receipt_cv":  cv.DateReceiptCv.Format(cf.FormatDateDisplay),
		"interview_method": cv.InterviewMethod,
		"contact_link":     cv.ContactLink,
		"media_id":         cv.MediaId,
		"media_id_other":         cv.MediaIdOther,
		"file_name":        cv.FileName,
		"file_content":     byteArr,
		"file_path": 		filePath,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get cv successful",
		Data:    resp,
	})
}

func (ctr *Controller) RemoveCv(c echo.Context) error {
	params := new(param.RemoveCvParam)
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

	count, err := ctr.RecruitmentRepo.CountCvById(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Cv does not exist",
		})
	}

	if err := ctr.RecruitmentRepo.DeleteCv(params.Id); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove cv successful",
	})
}

func (ctr *Controller) CreateCvComment(c echo.Context) error {
	params := new(param.CreateCvComment)
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

	rcColumns := []string{"assignees"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, rcColumns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	_, err = ctr.RecruitmentRepo.FindCvById(params.CvId)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Cv does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	user := c.Get("user_profile").(m.User)

	if !utils.FindIntInSlice(recruitment.Assignees, user.ID) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to create comment. Check assignees again",
		})
	}

	resp, err := ctr.RecruitmentRepo.InsertCvComment(
		user.OrganizationID,
		user.ID,
		params,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error: " + err.Error(),
		})
	}

	emailBody := user.UserProfile.FirstName + " " + user.UserProfile.LastName + " " + resp.Body
	if user.Organization.Email != "" && user.Organization.EmailPassword != "" {
		if err := ctr.NotificationRepo.InsertEmailNotiRequest(&m.EmailNotiRequest{
			NotiRequestId:  resp.NotiRequestId,
			OrganizationId: user.OrganizationID,
			Sender:         user.ID,
			ToUserIds:      resp.Assignees,
			Subject:        "【Notification】【Micro erp】Add a comment to cv",
			Content:        "Hi there, " + emailBody + ". Please click the button below for more information",
			Url:            os.Getenv("BASE_SPA_URL") + resp.Link,
			Template:       cf.Recruitment,
		}); err != nil {
			ctr.Logger.Error(err)
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create comment successful",
		Data: map[string]interface{}{
			"noti_request_id": resp.NotiRequestId,
		},
	})
}

func (ctr *Controller) EditCvComment(c echo.Context) error {
	params := new(param.EditCvComment)
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

	columns := []string{"created_by"}
	cvComment, err := ctr.RecruitmentRepo.SelectCvCommentById(params.Id, columns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Comment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.UserProfile.UserID != cvComment.CreatedBy {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to edit this comment",
		})
	}

	if err := ctr.RecruitmentRepo.UpdateCvComment(params.Id, params.Comment); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit comment successful",
	})
}

func (ctr *Controller) GetCvComments(c echo.Context) error {
	params := new(param.GetCvCommentsParam)
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

	rcColumns := []string{"assignees"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, rcColumns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID &&
		userProfile.RoleID != cf.ManagerRoleID &&
		!utils.FindIntInSlice(recruitment.Assignees, userProfile.UserProfile.UserID) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to get cvs",
		})
	}

	columns := []string{"id", "created_by", "comment", "created_at", "updated_at", "receivers"}
	cvComments, err := ctr.RecruitmentRepo.SelectCvCommentsByCvId(params.CvId, columns...)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var comments []map[string]interface{}
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	for _, cvComment := range cvComments {
		data := map[string]interface{}{
			"id":         cvComment.ID,
			"comment":    cvComment.Comment,
			"created_by": cvComment.CreatedBy,
			"receivers":  cvComment.Receivers,
			"created_at": cvComment.CreatedAt.In(location).Format(cf.FormatTimeDisplay),
			"updated_at": cvComment.UpdatedAt.In(location).Format(cf.FormatTimeDisplay),
		}

		comments = append(comments, data)
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			for _, assigneeID := range recruitment.Assignees {
				if user.UserID == assigneeID {
					users[user.UserID] = user.FullName
				}
			}
		}
	}

	dataResponse := map[string]interface{}{
		"comments": comments,
		"users":    users,
		"avatars":  ctr.SelectAssigneeAndAvatars(recruitment.Assignees),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get cv comments successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) GetLogCvStatus(c echo.Context) error {
	params := new(param.GetLogCvStatusParam)
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

	columns := []string{"status", "update_day"}
	logCvStates, err := ctr.RecruitmentRepo.SelectLogCvStates(params.CvId, columns...)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var dataResponse []map[string]interface{}
	for _, logcvstatus := range logCvStates {
		data := map[string]interface{}{
			"datetime_update": logcvstatus.UpdateDay.Format(cf.FormatTimeDisplay),
			"status":          logcvstatus.Status,
		}

		dataResponse = append(dataResponse, data)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get logs cv status successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) RemoveCvComment(c echo.Context) error {
	params := new(param.RemoveCvCommentParam)
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

	rcColumns := []string{"assignees"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, rcColumns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !utils.FindIntInSlice(recruitment.Assignees, userProfile.UserProfile.UserID) {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You do not have permission to remove comment",
		})
	}

	if err := ctr.RecruitmentRepo.DeleteCvCommentById(params.Id); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove comment successful",
	})
}

func (ctr *Controller) SelectAssigneeAndAvatars(assignees []int) map[int][]byte {
	records := make(map[int][]byte)
	if len(assignees) == 0 {
		return records
	}

	userIdAndAvatars, err := ctr.UserRepo.SelectAvatarUsers(assignees)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		panic(err)
	}

	for _, user := range userIdAndAvatars {
		var base64Img []byte
		if user.Avatar != "" {
			base64Img, err = ctr.Cloud.GetFileByFileName(user.Avatar, cf.AvatarFolderGCS)
			if err != nil {
				ctr.Logger.Error(err)
			}
		}

		records[user.UserId] = base64Img
	}

	return records
}

func (ctr *Controller) UploadDetailJob(c echo.Context) error {
	params := new(param.DetailJobParams)
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

	columns := []string{"job_name", "organization_id", "assignees"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentID, columns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCRUD_CV(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to do this action",
		})
	}

	directoryCloud := recruitment.GCSCVDirectory()
	millisecondTimeNow := int(time.Now().UTC().UnixNano() / int64(time.Millisecond))
	params.FileName = strings.Replace(strconv.Itoa(millisecondTimeNow)+"_"+params.FileName,
		" ", "_", -1)

	if err := ctr.Cloud.UploadFileToCloud(
		params.FileContent,
		params.FileName,
		directoryCloud,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	detailedRecruitment := m.Recruitment{
		BaseModel: cm.BaseModel{
			ID: params.RecruitmentID,
		},
		DetailJobFileName: params.FileName,
	}

	if err := ctr.RecruitmentRepo.UpdateDetailedJobFileName(&detailedRecruitment); err != nil {
		ctr.Logger.Errorf("Update detailed job got error: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when insert new detailed job recruitment",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Upload detail job successful",
	})
}

func (ctr *Controller) GetDetailJobFile(c echo.Context) error {
	params := new(param.GetDetailJobFile)
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

	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, []string{"organization_id", "job_name", "assignees", "detail_job_file_name"}...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if !recruitment.IsAllowCRUD_CV(userProfile) {
		return c.JSON(http.StatusForbidden, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You dont have permission to view detail of this CV",
		})
	}

	byteArr, err := ctr.Cloud.GetFileByFileName(recruitment.DetailJobFileName, recruitment.GCSCVDirectory())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when download CV from GCS",
		})
	}

	resp := map[string]interface{}{
		"file_content": byteArr,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get detail job file successful",
		Data:    resp,
	})
}

func (ctr *Controller) RemoveDetailJobFile(c echo.Context) error {
	params := new(param.RemoveDetailJobFile)
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

	rcColumns := []string{"job_name", "detail_job_file_name", "organization_id"}
	recruitment, err := ctr.RecruitmentRepo.SelectJob(params.RecruitmentId, rcColumns...)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Recruitment does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if err := ctr.Cloud.DeleteFileCloud(recruitment.DetailJobFileName, recruitment.GCSCVDirectory()); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when remove detailed job GCS",
		})
	}

	detailedRecruitment := m.Recruitment{
		BaseModel: cm.BaseModel{
			ID: params.RecruitmentId,
		},
		DetailJobFileName: "",
	}

	if err := ctr.RecruitmentRepo.UpdateDetailedJobFileName(&detailedRecruitment); err != nil {
		ctr.Logger.Errorf("Update detailed job got error: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error when update detailed job recruitment",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove detail job file successful",
	})
}

func (ctr *Controller) CreateLogCvStatus(c echo.Context) error {
	params := new(param.CreateLogCvStatus)
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

	t, err := time.Parse(cf.FormatDateNoSec, params.UpdateDay)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid value for field update_day",
		})
	}

	_, err = ctr.RecruitmentRepo.FindCvById(params.CvId)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Cv does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}
	userProfile := c.Get("user_profile").(m.User)

	logCvState := m.LogCvState{
		CvId:      params.CvId,
		Status:    params.Status,
		UpdateDay: t,
	}

	updateCv := m.Cv{
		BaseModel: cm.BaseModel{
			ID: params.CvId,
		},
		StatusCv: params.Status,
	}
	body, link, err := ctr.RecruitmentRepo.CreateLogCvStatusModel(&logCvState, userProfile.OrganizationID, userProfile.UserProfile.UserID, params, ctr.NotificationRepo)
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	registrationTokens, err := ctr.FcmTokenRepo.SelectMultiFcmTokens(params.Assignees, userProfile.UserProfile.UserID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
	if len(registrationTokens) > 0 {
		for _, token := range registrationTokens {
			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
			}
		}
	}

	if userProfile.Organization.Email != "" && userProfile.Organization.EmailPassword != "" {
		ctr.InitSmtp(userProfile.Organization.Email, userProfile.Organization.EmailPassword)
		emails, _ := ctr.UserRepo.SelectEmailByUserIds(params.Assignees)
		sampleData := new(param.SampleData)
		sampleData.SendTo = emails
		sampleData.Content = "Hi there, " + body + ". Please click the button below for more information"
		sampleData.URL = os.Getenv("BASE_SPA_URL") + link
		if err := ctr.SendMail(
			"【Notification】【Micro erp】Update cv status",
			sampleData,
			cf.Recruitment,
		); err != nil {
			ctr.Logger.Error(err)
		}
	}

	if err := ctr.RecruitmentRepo.UpdateCv(&updateCv); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System got error when update CV",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create log cv status successful",
	})
}
