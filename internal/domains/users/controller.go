package users

import (
	"encoding/base64"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	valid "github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"

	"github.com/go-pg/pg/v9"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	ex "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/excel"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type UserController struct {
	cm.BaseController
	email.SMTPGoMail

	UserRepo           rp.UserRepository
	OrgRepo            rp.OrgRepository
	UserProjectRepo    rp.UserProjectRepository
	LeaveRepo          rp.LeaveRepository
	ProjectRepo        rp.ProjectRepository
	EvaluationRepo     rp.EvaluationRepository
	BranchRepo         rp.BranchRepository
	JobTitleRepo       rp.JobTitleRepository
	TechnologyRepo     rp.TechnologyRepository
	UserTechnologyRepo rp.UserTechnologyRepository
	UserPermissionRepo rp.UserPermissionRepository
	cloud              gc.StorageUtility
}

func NewUserController(
	logger echo.Logger,
	userRepo rp.UserRepository,
	orgRepo rp.OrgRepository,
	userProjectRepo rp.UserProjectRepository,
	leaveRepo rp.LeaveRepository,
	projectRepo rp.ProjectRepository,
	evaluationRepo rp.EvaluationRepository,
	branchRepo rp.BranchRepository,
	jobTitleRepo rp.JobTitleRepository,
	technologyRepo rp.TechnologyRepository,
	userTechnologyRepo rp.UserTechnologyRepository,
	userPermissionRepo rp.UserPermissionRepository,
	cloud gc.StorageUtility,
) (ctr *UserController) {
	ctr = &UserController{
		cm.BaseController{}, email.SMTPGoMail{},
		userRepo, orgRepo,
		userProjectRepo, leaveRepo,
		projectRepo, evaluationRepo,
		branchRepo, jobTitleRepo,
		technologyRepo, userTechnologyRepo,
		userPermissionRepo, cloud,
	}
	ctr.Init(logger)
	return
}

// GetLoginUser : get information user login
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) GetLoginUser(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["id"].(float64)

	user, err := ctr.UserRepo.GetUserProfile(int(userID))
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	rankLogRecords, err := ctr.UserRepo.GetUserRankLogByUserID(int(userID))
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User rank logs is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var userRankLogs []map[string]interface{}
	for _, record := range rankLogRecords {
		log := map[string]interface{}{
			"rank":       record.Rank,
			"created_at": record.CreatedAt.Format(cf.FormatDateDatabase),
		}

		userRankLogs = append(userRankLogs, log)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(user.OrganizationID)
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

	functionRecords, err := ctr.UserPermissionRepo.SelectPermissions(user.OrganizationID, user.ID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.SuccessResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	functionPermission := make(map[int]interface{})

	for i := 1; i <= cf.TOTALMODULE; i++ {
		var module (map[string]interface{})
		var modules []map[string]interface{}

		for _, record := range functionRecords {
			if record.ModuleId == i {
				module = map[string]interface{}{
					"function_id": record.FunctionId,
					"status":      record.Status,
				}

				modules = append(modules, module)
				functionPermission[record.ModuleId] = modules
			}
		}
	}

	if user.RoleID == cf.GeneralManagerRoleID {
		records, err := ctr.LeaveRepo.CountExpireDate(user.OrganizationID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		var leaveBonusParams []param.LeaveBonus
		if len(records) > 0 {
			for _, record := range records {
				leaveBonusParam := new(param.LeaveBonus)
				leaveBonusParam.UserID = record.UserId
				leaveBonusParam.LeaveBonusTypeID = cf.ClearLeave
				leaveBonusParam.YearBelong = time.Now().Year()
				leaveBonusParam.Reason = "Clear expiration date"
				leaveBonusParam.Hour = -record.TotalHour
				leaveBonusParam.HourRemaining = 0

				leaveBonusParams = append(leaveBonusParams, *leaveBonusParam)
			}

			if len(leaveBonusParams) > 0 {
				if err := ctr.LeaveRepo.InsertLeaveBonusWithTx(
					user.OrganizationID,
					user.ID,
					&leaveBonusParams,
				); err != nil {
					ctr.Logger.Error(err)
				}
			}
		}

		if err := ctr.LeaveRepo.ClearExpireDate(user.OrganizationID); err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}
	}

	dataResponse := map[string]interface{}{
		"id":                       user.ID,
		"email":                    user.Email,
		"language_id":              user.LanguageId,
		"phone_number":             user.UserProfile.PhoneNumber,
		"first_name":               user.UserProfile.FirstName,
		"last_name":                user.UserProfile.LastName,
		"avatar":                   user.UserProfile.Avatar,
		"birthday":                 user.UserProfile.Birthday,
		"branch":                   user.UserProfile.Branch,
		"role_id":                  user.RoleID,
		"role_name":                user.Role.Name,
		"organization_id":          user.OrganizationID,
		"organization_name":        user.Organization.Name,
		"user_rank_logs":           userRankLogs,
		"branch_list_box":          branches,
		"setting_step":             user.Organization.SettingStep,
		"expiration_reset_day_off": user.Organization.ExpirationResetDayOff,
		"func_permission":          functionPermission,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    dataResponse,
	})
}

// ForgotPassword : call model checkCode() to check code valid
// Params                :
// Returns               : return data with struct JsonResponse
func (ctr *UserController) ForgotPassword(c echo.Context) error {
	forgotPasswordParams := new(param.ForgotPassParams)
	if err := c.Bind(forgotPasswordParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	objUser, err := ctr.UserRepo.GetUserByEmailOrganizationID(forgotPasswordParams.Email, forgotPasswordParams.OrganizationID)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Email is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// check expired code update password
	if objUser.CodeExpiredAt.After(utils.TimeNowUTC()) {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Password reset email was sent please check your email and spam or wait 2 hour to use this function.",
		})
	}

	org, err := ctr.OrgRepo.SelectEmailAndPassword(forgotPasswordParams.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}
	if org.Email == "" || org.EmailPassword == "" {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The email system is currently down. Please try later",
		})
	}

	timeNow := utils.TimeNowUTC()
	// code expired after 2 hours
	timeExpired := timeNow.Add(cf.ExpiredHours * time.Hour)

	// create secret code with random stirng and time now millisecond
	codeForgotPass := utils.GetUniqueString()
	hashCodeForgotPass := utils.GetSHA256Hash(codeForgotPass)

	var setForgotPassParams = param.SetForgotPassParams{
		UserID:            objUser.ID,
		ResetPasswordCode: hashCodeForgotPass,
		CodeExpiredAt:     timeExpired,
	}

	err = ctr.UserRepo.SetForgotPassParams(setForgotPassParams)
	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	ctr.InitSmtp(org.Email, org.EmailPassword)

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{objUser.Email}
	sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/ResetPassword/" + codeForgotPass

	if err := ctr.SendMail("Micro Erp Registration Email", sampleData, cf.TemplateMailForgotPass); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
	})
}

// CheckResetPasswordCode :
// Params  :
// Returns :
func (ctr *UserController) CheckResetPasswordCode(c echo.Context) error {
	checkResetCodeParams := new(param.CheckResetCodeParams)

	if err := c.Bind(checkResetCodeParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	hashCode := utils.GetSHA256Hash(checkResetCodeParams.ResetPasswordCode)
	user, err := ctr.UserRepo.GetUserByResetCode(hashCode)
	if err != nil && err.Error() == pg.ErrNoRows.Error() {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your reset password link is wrong.",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if user.CodeExpiredAt.Before(utils.TimeNowUTC()) {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your reset password link is expired.",
		})
	}

	dataResponse := map[string]interface{}{
		"user_id":             user.ID,
		"email":               user.Email,
		"organization_id":     user.OrganizationID,
		"reset_password_code": user.ResetPasswordCode,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "You can reset password now",
		Data:    dataResponse,
	})
}

// ResetPassword :
// Params  :
// Returns :
func (ctr *UserController) ResetPassword(c echo.Context) error {
	resetPasswordParams := new(param.ResetPasswordParams)

	if err := c.Bind(resetPasswordParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	err := ctr.UserRepo.UpdateResetPassword(resetPasswordParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// -------------------------------------> get organization info to redirect login page <-----------------------------------
	orgObj, err := ctr.OrgRepo.GetOrganizationByID(resetPasswordParams.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	dataResponse := map[string]interface{}{
		"user_id":           resetPasswordParams.UserID,
		"organization_id":   orgObj.ID,
		"organization_name": orgObj.Name,
		"organization_tag":  orgObj.Tag,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Password Successful reset",
		Data:    dataResponse,
	})
}

// InsertNewEmailForUpdate : save new code for update email
// Params  :
// Returns : return data with struct JsonResponse
func (ctr *UserController) InsertNewEmailForUpdate(c echo.Context) error {
	updateEmailParams := new(param.UpdateEmailParams)

	if err := c.Bind(updateEmailParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	_, err := valid.ValidateStruct(updateEmailParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	if userProfile.Email == updateEmailParams.EmailForUpdate {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Update email same as old email.",
		})
	}

	checkEmailOrgObj, err := ctr.UserRepo.GetUserByEmailOrganizationID(updateEmailParams.EmailForUpdate, userProfile.OrganizationID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if checkEmailOrgObj.ID != 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Email already been used.",
		})
	}

	org, err := ctr.OrgRepo.SelectEmailAndPassword(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}
	if org.Email == "" || org.EmailPassword == "" {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your organization must have an email",
		})
	}

	timeNow := utils.TimeNowUTC()
	// code expired after 2 hours
	expiredTime := timeNow.Add(cf.ExpiredHours * time.Hour)

	// create secret code with random stirng and time now millisecond
	code := utils.GetUniqueString()
	hashCode := utils.GetSHA256Hash(code)

	var setUpdateEmailParams = param.SetUpdateEmailParams{
		EmailForUpdate:               updateEmailParams.EmailForUpdate,
		UpdateEmailCode:              hashCode,
		UpdateEmailCodeCodeExpiredAt: expiredTime,
	}

	err = ctr.UserRepo.SetNewUpdateEmail(setUpdateEmailParams, userProfile.ID)
	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	ctr.InitSmtp(org.Email, org.EmailPassword)

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{updateEmailParams.EmailForUpdate}
	sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/ChangeEmail/" + code
	if err := ctr.SendMail("Micro Erp Change Email", sampleData, cf.ChangeEmailTemplate); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Email update request was sent. Please check update mail to complete",
	})
}

// ChangeEmail : check valid change email code
// Params   :
// Returns : return data with struct JsonResponse
func (ctr *UserController) ChangeEmail(c echo.Context) error {
	checkChangeEmailCodeParams := new(param.CheckChangeEmailCodeParams)

	if err := c.Bind(checkChangeEmailCodeParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	_, err := valid.ValidateStruct(checkChangeEmailCodeParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	hashCode := utils.GetSHA256Hash(checkChangeEmailCodeParams.ChangeEmailCode)
	user, err := ctr.UserRepo.GetUserByChangeEmailCode(hashCode)

	if err != nil && err.Error() == pg.ErrNoRows.Error() {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your change email link is wrong.",
		})
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if user.UpdateEmailCodeExpiredAt.Before(utils.TimeNowUTC()) {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your change email link is expired.",
		})
	}

	checkEmailOrgObj, err := ctr.UserRepo.GetUserByEmailOrganizationID(user.EmailForUpdate, user.OrganizationID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if checkEmailOrgObj.ID != 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Email already been used.",
		})
	}

	err = ctr.UserRepo.UpdateEmail(user)

	dataResponse := map[string]interface{}{
		"user_id":           user.ID,
		"email":             user.EmailForUpdate,
		"organization_id":   user.OrganizationID,
		"change_email_code": user.UpdateEmailCode,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Change Email Success",
		Data:    dataResponse,
	})
}

// ChangePassword : change password (user account page)
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) ChangePassword(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	changePasswordParams := new(param.ChangePasswordParams)

	if err := c.Bind(changePasswordParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	_, err := valid.ValidateStruct(changePasswordParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	if strings.Compare(changePasswordParams.NewPassword, changePasswordParams.RepeatNewPassword) != 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Repeat new password does not match",
		})
	}

	checkCurrentPass, err := ctr.UserRepo.CheckCurrentPassword(userProfile.ID, changePasswordParams.CurrentPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if !checkCurrentPass {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Current password not correct",
		})
	}

	err = ctr.UserRepo.UpdatePasswordByID(userProfile.ID, changePasswordParams.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Change password successful",
	})
}

// UpdateProfile : get edit user
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) UpdateProfile(c echo.Context) error {
	editor := c.Get("user_profile").(m.User)

	editProfileParams := new(param.EditProfileParams)

	if err := c.Bind(editProfileParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	_, err := valid.ValidateStruct(editProfileParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	if editProfileParams.DateSeverance != "" {
		_, err := time.Parse(cf.FormatDateDatabase, editProfileParams.DateSeverance)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "The value of date_severance is invalid",
			})
		}
	}

	editUserProfile, err := ctr.UserRepo.GetUserProfileExpand(editProfileParams.UserID)

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

	if editor.OrganizationID != editUserProfile.OrganizationID {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "User profile not found",
		})
	}

	if editor.ID != editUserProfile.ID && editor.RoleID != cf.GeneralManagerRoleID && editor.RoleID != cf.ManagerRoleID {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You don't have permission to edit this profile",
		})
	}

	// upload avatar
	if editProfileParams.FlagEditBasicProfile && editProfileParams.FlagEditAvatar && editProfileParams.Avatar != "" {
		readerImage := base64.NewDecoder(base64.StdEncoding, strings.NewReader(editProfileParams.Avatar))
		_, formatImageAvatar, err := image.DecodeConfig(readerImage)

		if err != nil {
			ctr.Logger.Error(err)
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Upload avatar error",
			})
		}

		if _, check := utils.FindStringInArray(cf.AllowFormatImageList, formatImageAvatar); !check {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "The Avatar field must be an image",
			})
		}

		millisecondTimeNow := int(time.Now().UnixNano() / int64(time.Millisecond))
		nameNewAvatar := strconv.Itoa(editProfileParams.UserID) + "_" + strconv.Itoa(millisecondTimeNow) + "." + formatImageAvatar
		err = ctr.cloud.UploadFileToCloud(editProfileParams.Avatar, nameNewAvatar, cf.AvatarFolderGCS)

		if err != nil {
			ctr.Logger.Error(err)

			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Upload avatar error",
			})
		}

		editProfileParams.Avatar = nameNewAvatar
	}
	// upload avatar

	editProfileParams.EditorRoleID = editor.RoleID
	err = ctr.UserRepo.SaveProfile(editProfileParams)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates") {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Duplicate employee id value violates",
		})
	}

	if err != nil {
		// if strings.Contains(err.Error(), "duplicate key value violates") {
		// 	return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
		// 		Status:  cf.FailResponseCode,
		// 		Message: "Duplicate employee id value violates",
		// 	})
		// }
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if editUserProfile.Rank != editProfileParams.Rank {
		err = ctr.UserRepo.InsertIntoUserRankLog(editProfileParams.UserID, editProfileParams.Rank)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	// remove old avatar from cloud
	if editProfileParams.FlagEditBasicProfile && editProfileParams.FlagEditAvatar && editUserProfile.Avatar != "" {
		ctr.cloud.DeleteFileCloud(editUserProfile.Avatar, cf.AvatarFolderGCS)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit profile successful",
	})
}

// GetUserInfo : get user profile
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) GetUserInfo(c echo.Context) error {
	viewerProfile := c.Get("user_profile").(m.User)

	user := m.UserProfileExpand{}
	userInfoParams := new(param.UserInfoParams)

	if err := c.Bind(userInfoParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	_, err := valid.ValidateStruct(userInfoParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	user, err = ctr.UserRepo.GetUserProfileExpand(userInfoParams.UserID)

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

	if viewerProfile.OrganizationID != user.OrganizationID {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "User profile not found",
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

	// check Birthday nil return string nil
	strBirthday := ""
	if !user.Birthday.IsZero() {
		strBirthday = user.Birthday.Format(cf.FormatDateDisplay)
	}

	strCompanyJoinedDate := ""
	if !user.CompanyJoinedDate.IsZero() {
		strCompanyJoinedDate = user.CompanyJoinedDate.Format(cf.FormatDateDisplay)
	}

	dateOfIdentityCard := ""
	if !user.DateOfIdentityCard.IsZero() {
		dateOfIdentityCard = user.DateOfIdentityCard.Format(cf.FormatDateDisplay)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(viewerProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var branchName string
	if user.Branch > 0 && len(branchRecords) > 0 {
		for _, record := range branchRecords {
			if record.Id == user.Branch {
				branchName = record.Name
			}
		}
	}

	jobTitleRecords, err := ctr.JobTitleRepo.SelectJobTitles(user.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var jobTitleName string
	if user.JobTitle > 0 && len(jobTitleRecords) > 0 {
		for _, record := range jobTitleRecords {
			if record.Id == user.JobTitle {
				jobTitleName = record.Name
			}
		}
	}

	dataResponse := map[string]interface{}{
		"user_id":               user.ID,
		"email":                 user.Email,
		"phone_number":          user.PhoneNumber,
		"first_name":            user.FirstName,
		"last_name":             user.LastName,
		"avatar":                base64Img,
		"birthday":              strBirthday,
		"role_id":               user.RoleID,
		"role_name":             user.RoleName,
		"organization_id":       user.OrganizationID,
		"organization_name":     user.OrganizationName,
		"rank":                  user.Rank,
		"rank_name":             cf.RankList[user.Rank],
		"job_title":             user.JobTitle,
		"job_title_name":        jobTitleName,
		"company_joined_date":   strCompanyJoinedDate,
		"skill":                 user.Skill,
		"language":              user.Language,
		"education":             user.Education,
		"award":                 user.Award,
		"experience":            user.Experience,
		"introduce":             user.Introduce,
		"branch":                user.Branch,
		"branch_name":           branchName,
		"employee_id":           user.EmployeeId,
		"current_address":       user.CurrentAddress,
		"tax_code":              user.TaxCode,
		"status":                user.Status,
		"identity_card":         user.IdentityCard,
		"occupation":            user.Occupation,
		"country":               user.Country,
		"work_place":            user.WorkPlace,
		"job_position":			user.JobPosition,
		"place_of_birth":		 user.PlaceOfBirth,
		"gender":				 user.Gender,
		"marital_status":		 user.MaritalStatus,
		"book_number_bhxh":		 user.BookNumberBhxh,
		"account_number_vcb":		user.AccountNumberVcb,
		"nation":				 user.Nation,
		"religion":				 user.Religion,
		"name_of_emergency":		user.NameOfEmergency,
		"relationships_of_emergency":		user.RelationshipsOfEmergency,
		"address_of_emergency":		user.AddressOfEmergency,
		"license_plates":		 user.LicensePlates,
		"permanent_residence":   user.PermanentResidence,
		"date_of_identity_card": dateOfIdentityCard,
		"user_birth_place":      user.UserBirthPlace,
		"id_type":               user.IDType,
		"place_of_issue":        user.PlaceOfIssue,
		"department": func(department *string) string {
			if department != nil {
				return *department
			}

			return ""
		}(user.Department),
		"date_severance": func(dateSeverance *time.Time) string {
			if dateSeverance != nil {
				return dateSeverance.Format(cf.FormatDateDisplay)
			}

			return ""
		}(user.DateSeverance),
		"reasons_severance": func(reasonsSeverance *string) string {
			if reasonsSeverance != nil {
				return *reasonsSeverance
			}

			return ""
		}(user.ReasonsSeverance),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get profile successful",
		Data:    dataResponse,
	})
}

// GetListItemProfile : get list ranking, branch, language...
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) GetListItemProfile(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
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

	jobTitleRecords, err := ctr.JobTitleRepo.SelectJobTitles(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	jobTitles := make(map[int]string)
	if len(jobTitleRecords) > 0 {
		for _, record := range jobTitleRecords {
			jobTitles[record.Id] = record.Name
		}
	}

	technologyRecords, err := ctr.TechnologyRepo.SelectTechnologies(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	technologies := make(map[int]string)
	if len(technologyRecords) > 0 {
		for _, record := range technologyRecords {
			technologies[record.Id] = record.Name
		}
	}

	dataResponse := map[string]interface{}{
		"rank_list":           cf.RankList,
		"branch_list":         branches,
		"job_title_list":      jobTitles,
		"language_list":       cf.LanguageList,
		"level_language_list": cf.LevelLanguageList,
		"technology_list":     technologies,
		"level_skill_list":    cf.LevelSkillList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get profile successful",
		Data:    dataResponse,
	})
}

// GetAvatarUserLogin : get avatar base 64...
// Params  : echo.Context
// Returns : JSON
func (ctr *UserController) GetAvatarUserLogin(c echo.Context) error {
	userInfo := c.Get("user_profile").(m.User)
	dataResponse := map[string]interface{}{}

	if userInfo.UserProfile.Avatar == "" {
		dataResponse["avatar"] = nil
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.SuccessResponseCode,
			Message: "User has no avatar",
			Data:    dataResponse,
		})
	}

	base64Img, err := ctr.cloud.GetFileByFileName(userInfo.UserProfile.Avatar, cf.AvatarFolderGCS)

	if err != nil {
		ctr.Logger.Error(err)
		base64Img = nil
	}

	dataResponse["avatar"] = base64Img

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get avatar successful",
		Data:    dataResponse,
	})
}

func (ctr *UserController) DisplayLanguageSetting(c echo.Context) error {
	languageSettingParams := new(param.LanguageSettingParams)
	if err := c.Bind(languageSettingParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid params",
		})
	}

	if _, err := valid.ValidateStruct(languageSettingParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err := ctr.UserRepo.UpdateLanguageSetting(userProfile.OrganizationID, languageSettingParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Language setting successful",
	})
}

// SearchUserProfile : search, filter, order profile list
// Params            : echo.Context
// Returns           : return data with struct JsonResponse
func (ctr *UserController) SearchUserProfile(c echo.Context) error {
	userProfileListParams := new(param.UserProfileListParams)
	userProfileListParams.RowPerPage = cf.RowPerPageProfileList

	if err := c.Bind(userProfileListParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid params",
		})
	}

	// check field type date have format format 2006/01/02 (YYYY/mm/dd)
	valid.TagMap["formatDisplayDate"] = utils.ValidatorFormatDisplayDate()

	_, err := valid.ValidateStruct(userProfileListParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	// return error if date from after date to
	if !userProfileListParams.DateFrom.IsZero() && !userProfileListParams.DateTo.IsZero() {
		if userProfileListParams.DateFrom.After(userProfileListParams.DateTo) {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Date from should be less than or equal date to",
			})
		}
	}

	userProfile := c.Get("user_profile").(m.User)
	userProfileList, totalRow, err := ctr.UserRepo.GetUserProfileList(userProfile.OrganizationID, userProfileListParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// user := m.UserProfileExpand{}
	profileList := []map[string]interface{}{}
	if userProfileList != nil {
		for _, item := range userProfileList {
			strCompanyJoinedDate := ""
			if !item.UserProfile.CompanyJoinedDate.IsZero() {
				strCompanyJoinedDate = item.UserProfile.CompanyJoinedDate.Format(cf.FormatDateDisplay)
			}

			var base64Img []byte = nil
			if item.UserProfile.Avatar != "" {
				base64Img, err = ctr.cloud.GetFileByFileName(item.UserProfile.Avatar, cf.AvatarFolderGCS)

				if err != nil {
					ctr.Logger.Error(err)
					base64Img = nil
				}
			}

			profileObj := map[string]interface{}{
				"id":                item.ID,
				"first_name":        item.UserProfile.FirstName,
				"last_name":         item.UserProfile.LastName,
				"email":             item.Email,
				"company_join_date": strCompanyJoinedDate,
				"branch":            item.UserProfile.UserBranch.Name,
				"role":              item.Role.Name,
				"phone_number":		item.UserProfile.PhoneNumber,
				"avatar":            base64Img,
				"job_position":		item.UserProfile.JobPosition,
				"job_title":         item.UserProfile.JobTitle,
				"status":            item.UserProfile.Status,
			}
			profileList = append(profileList, profileObj)
		}
	}

	pagination := map[string]interface{}{
		"current_page": userProfileListParams.CurrentPage,
		"total_row":    totalRow,
		"row_perpage":  userProfileListParams.RowPerPage,
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

	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil {
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
			"pagination":        pagination,
			"profile_list":      profileList,
			"branch_select_box": branches,
			"rank_select_box":   cf.RankList,
			"job_title_list":	cf.JobTitleExcel,
			"working_status":  cf.WorkingStatus,
			"users":             userList,
		},
	})
}

// Statistic : Statistic system
func (ctr *UserController) Statistic(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	rankLogRecords, err := ctr.UserRepo.GetUserRankLogByUserID(userProfile.UserProfile.UserID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User rank logs is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var userRankLogs []map[string]interface{}
	for _, record := range rankLogRecords {
		log := map[string]interface{}{
			"rank":       record.Rank,
			"created_at": record.CreatedAt.Format(cf.FormatDateDatabase),
		}

		userRankLogs = append(userRankLogs, log)
	}

	numberPeopleBranch, err := ctr.UserRepo.CountPeopleBranch(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	numberPeopleJobTitle, err := ctr.UserRepo.CountPeopleJobTitle(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	numberPeopleJapaneseLevel, err := ctr.UserRepo.CountJapaneseLevel(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	numberPeopleInterestTechnology, err := ctr.UserTechnologyRepo.CountUserInterestTechnology(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	evaluationRankRecords, err := ctr.EvaluationRepo.EvaluationRankLastFourQuarter(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var evaluationRankResponse param.EvaluationRankDatasets
	for _, rank := range evaluationRankRecords {
		datetime := "Q" + strconv.Itoa(rank.Quarter) + "-" + strconv.Itoa(rank.Year)
		if !ctr.checkExistString(evaluationRankResponse.Datetime, datetime) {
			evaluationRankResponse.Datetime = append(evaluationRankResponse.Datetime, datetime)
		}
	}

	var keys []int
	for k, _ := range cf.EvaluationRankList {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	for _, key := range keys {
		data := ctr.getEvaluationRankData(evaluationRankRecords, key)
		dataset := param.Dataset{
			Rank: cf.EvaluationRankList[key],
			Data: data,
		}
		evaluationRankResponse.Datasets = append(evaluationRankResponse.Datasets, dataset)
	}

	dayUsed, _, dayRemaining, _ := ctr.LeaveRepo.GetLeaveDayStatus(userProfile.OrganizationID, userProfile.UserProfile.UserID, time.Now().Year())
	dayOffInfo := map[string]float64{
		"day_used":      dayUsed,
		"day_remaining": dayRemaining,
	}

	totalUser, err := ctr.UserRepo.CountUser(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Users is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	totalProject, err := ctr.ProjectRepo.CountProject(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Projects is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	total := map[string]int{
		"users":    totalUser,
		"projects": totalProject,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"user_rank_logs":                    userRankLogs,
			"number_people_branch":              numberPeopleBranch,
			"number_people_job_title":           numberPeopleJobTitle,
			"number_people_japanese_level":      numberPeopleJapaneseLevel,
			"number_people_interest_technology": numberPeopleInterestTechnology,
			"evaluation_rank":                   evaluationRankResponse,
			"day_off_info":                      dayOffInfo,
			"total":                             total,
		},
	})
}

func (ctr *UserController) ImportProfiles(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	extensionFile := filepath.Ext(strings.TrimSpace(file.Filename))
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

	userProfile := c.Get("user_profile").(m.User)
	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}
	if len(branchRecords) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Branches is not exist",
		})
	}
	var branches []int
	for _, record := range branchRecords {
		branches = append(branches, record.Id)
	}

	jobTitleRecords, err := ctr.JobTitleRepo.SelectJobTitles(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}
	if len(jobTitleRecords) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Job titles is not exist",
		})
	}
	var jobTitles []int
	for _, record := range jobTitleRecords {
		jobTitles = append(jobTitles, record.Id)
	}

	var updateProfileParams []param.UpdateProfileParams
	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) > 7 {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid Params",
				Data:    i + 1,
			})
		}

		record := [7]string{"", "", "", "", "", "", ""}
		for j := 0; j < len(row); j++ {
			record[j] = row[j]
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

		_, err := ctr.UserRepo.SelectUserIdByEmployeeId(userProfile.OrganizationID, record[0])
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

		updateProfileParam := new(param.UpdateProfileParams)
		updateProfileParam.EmployeeId = record[0]

		branch, err := strconv.Atoi(record[6])
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}
		if !ctr.checkExistInt(branches, branch) {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}
		updateProfileParam.Branch = branch

		jobTitle, err := strconv.Atoi(record[3])
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}
		if !ctr.checkExistInt(jobTitles, jobTitle) {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}
		updateProfileParam.JobTitle = jobTitle

		if extensionFile == ".xlsx" {
			updateProfileParam.Birthday, err = time.Parse(cf.FormatDateDatabase, ex.ConvertExcelDate(record[1]))
		} else {
			updateProfileParam.Birthday, err = time.Parse(cf.FormatDateDatabase, record[1])
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		updateProfileParam.Rank, err = strconv.Atoi(record[2])
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		if extensionFile == ".xlsx" {
			updateProfileParam.CompanyJoinedDate, err = time.Parse(
				cf.FormatDateDatabase,
				ex.ConvertExcelDate(record[5]))
		} else {
			updateProfileParam.CompanyJoinedDate, err = time.Parse(cf.FormatDateDatabase, record[5])
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    i + 1,
			})
		}

		updateProfileParam.PhoneNumber = "0" + record[4]
		updateProfileParams = append(updateProfileParams, *updateProfileParam)
	}

	err = ctr.UserRepo.UpdateProfileWithTx(&updateProfileParams)
	if err != nil {
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

func (ctr *UserController) DownloadTemplate(c echo.Context) error {
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
			return c.Attachment("internal/platform/excel/template/import-profile.xlsx", "import-profile.xlsx")
		} else if userProfile.LanguageId == cf.VnLanguageId {
			return c.Attachment("internal/platform/excel/template/import-profile-vn.xlsx", "import-profile-vn.xlsx")
		} else {
			return c.Attachment("internal/platform/excel/template/import-profile-jp.xlsx", "import-profile-jp.xlsx")
		}
	}

	if userProfile.LanguageId == cf.EnLanguageId {
		return c.Attachment("internal/platform/excel/template/import-profile.csv", "import-profile.csv")
	} else if userProfile.LanguageId == cf.VnLanguageId {
		return c.Attachment("internal/platform/excel/template/import-profile-vn.csv", "import-profile-vn.csv")
	} else {
		return c.Attachment("internal/platform/excel/template/import-profile-jp.csv", "import-profile-jp.csv")
	}
}

func (ctr *UserController) JpLevelStatisticDetail(c echo.Context) error {
	jpLevelStatisticDetailParams := new(param.JpLevelStatisticDetailParams)
	if err := c.Bind(jpLevelStatisticDetailParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(jpLevelStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, totalRow, err := ctr.UserRepo.SelectUsersByJpLevel(userProfile.OrganizationID, jpLevelStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": jpLevelStatisticDetailParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": jpLevelStatisticDetailParams.RowPerPage,
	}

	var responses []map[string]interface{}
	for _, record := range records {
		res := map[string]interface{}{
			"user_id":             record.UserId,
			"full_name":           record.FullName,
			"job_title":           record.JobTitle,
			"branch":              record.Branch,
			"birthday":            record.Birthday.Format(cf.FormatDateDisplay),
			"company_joined_date": record.CompanyJoinedDate.Format(cf.FormatDateDisplay),
		}
		responses = append(responses, res)
	}

	dataResponse := map[string]interface{}{
		"pagination":       pagination,
		"statistic_detail": responses,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get japanese level statistic detail successful",
		Data:    dataResponse,
	})
}

func (ctr *UserController) checkExistString(arr []string, elm string) bool {
	if len(arr) == 0 {
		return false
	}

	for _, e := range arr {
		if e == elm {
			return true
		}
	}
	return false
}

func (ctr *UserController) checkExistInt(arr []int, elm int) bool {
	if len(arr) == 0 {
		return false
	}

	for _, e := range arr {
		if e == elm {
			return true
		}
	}
	return false
}

func (ctr *UserController) getEvaluationRankData(arr []param.RankLastFourQuarter, elm int) [4]int {
	data := [4]int{0, 0, 0, 0}
	for _, rank := range arr {
		if rank.Rank == elm {
			data[rank.Quarter-1] = rank.Amount
		}
	}

	return data
}
func (ctr *UserController) ExportExcel(c echo.Context) error {
	params := new(param.UserProfileListParams)
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
	userProfileList, _, err := ctr.UserRepo.GetUserProfileList(userProfile.OrganizationID, params)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "A", 35)
	_ = f.SetColWidth("Sheet1", "B", "B", 20)
	_ = f.SetColWidth("Sheet1", "C", "C", 30)
	_ = f.SetColWidth("Sheet1", "F", "F", 30)
	_ = f.SetColWidth("Sheet1", "C", "C", 35)
	_ = f.SetColWidth("Sheet1", "G", "G", 40)
	_ = f.SetColWidth("Sheet1", "D", "E", 25)
	_ = f.SetColWidth("Sheet1", "H", "I", 20)

	titleStyle, _ := f.NewStyle(`{
		"font":{"bold":true, "size":16},
		"alignment":{"horizontal":"center", "vertical":"center"}
	}`)
	contentStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center", "vertical":"center"}}`)

	var categoriesByLanguage map[string]string
	// var dayOfWeekByLanguage map[string]string
	// var overtimeTypeLanguage map[int]string
	// var mapStatusOvertimeType map[int]string

	if userProfile.LanguageId == cf.EnLanguageId {
		categoriesByLanguage = cf.EnUserCategories
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categoriesByLanguage = cf.VnUserCategories
	} else {
		categoriesByLanguage = cf.JpUserCategories
	}

	categories := map[string]string{
		"A1": categoriesByLanguage["Full Name"],
		"B1": categoriesByLanguage["Birthday"],
		"C1": categoriesByLanguage["Company join date"],
		"D1": categoriesByLanguage["Employee Id"],
		"E1": categoriesByLanguage["Phone number"],
		"F1": categoriesByLanguage["Job title"],
		"G1": categoriesByLanguage["Email"],
		"H1": categoriesByLanguage["Branch"],
		"I1": categoriesByLanguage["Gender"],
	}
	for k, v := range categories {
		_ = f.SetCellValue("Sheet1", k, v)
	}
	_ = f.SetCellStyle("Sheet1", "A1", "M1", titleStyle)
	_ = f.SetColStyle("Sheet1", "A", contentStyle)
	_ = f.SetColStyle("Sheet1", "B", contentStyle)
	_ = f.SetColStyle("Sheet1", "C", contentStyle)
	_ = f.SetColStyle("Sheet1", "D", contentStyle)
	_ = f.SetColStyle("Sheet1", "E", contentStyle)
	_ = f.SetColStyle("Sheet1", "F", contentStyle)
	_ = f.SetColStyle("Sheet1", "G", contentStyle)
	_ = f.SetColStyle("Sheet1", "H", contentStyle)
	_ = f.SetColStyle("Sheet1", "I", contentStyle)

	idx := 2
	for i, item := range userProfileList {
		pos := idx + i
		birthday := item.UserProfile.Birthday.Format("2006/01/02")
		joinDate := item.UserProfile.CompanyJoinedDate.Format("2006/01/02")
		values := map[string]interface{}{
			"A" + strconv.Itoa(pos): item.UserProfile.FirstName + " " + item.UserProfile.LastName,
			"B" + strconv.Itoa(pos): birthday,
			"C" + strconv.Itoa(pos): joinDate,
			"D" + strconv.Itoa(pos): item.UserProfile.EmployeeId,
			"E" + strconv.Itoa(pos): item.UserProfile.PhoneNumber,
			"F" + strconv.Itoa(pos): cf.JobTitleExcel[item.UserProfile.JobTitle],
			"G" + strconv.Itoa(pos): item.Email,
			"H" + strconv.Itoa(pos): cf.BranchExcel[item.UserProfile.Branch],
			"I" + strconv.Itoa(pos): cf.GenderExcel[item.UserProfile.Gender],
			
		}

		for k, v := range values {
			_ = f.SetCellValue("Sheet1", k, v)
		}
	}

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}
