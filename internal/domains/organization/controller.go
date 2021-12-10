package organization

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	cr "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cron"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type OrgController struct {
	cm.BaseController
	email.SMTPGoMail

	UserRepo    rp.UserRepository
	RegCodeRepo rp.RegCodeRepository
	OrgRepo     rp.OrgRepository
	AdminRepo   rp.AdminRepository
	RequestRepo rp.RegistRequestRepository
	cr.EtrCron
}

func NewOrgController(
	logger echo.Logger,
	userRepo rp.UserRepository,
	regCodeRepo rp.RegCodeRepository,
	orgRepo rp.OrgRepository,
	requestRepo rp.RegistRequestRepository,
	adminRepo rp.AdminRepository) (ctr *OrgController) {
	ctr = &OrgController{cm.BaseController{}, email.SMTPGoMail{}, userRepo, regCodeRepo, orgRepo, adminRepo, requestRepo, cr.EtrCron{}}
	ctr.Init(logger)
	ctr.InitCron("Asia/Ho_Chi_Minh")

	return
}

// FindOrganization  : find organization by tag
// Params  : echo.Context
// Returns : information organization(JSON)
func (ctr *OrgController) FindOrganization(c echo.Context) error {
	orgTag := strings.ToUpper(c.FormValue("tag_organization"))

	if !valid.IsAlphanumeric(orgTag) {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid params",
		})
	}

	org, err := ctr.OrgRepo.FindOrganizationByTag(orgTag)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Organization is not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	dataResponse := map[string]interface{}{
		"id":    org.ID,
		"tag":   org.Tag,
		"name":  org.Name,
		"asas":  os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		"asas1": os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    dataResponse,
	})
}

// SelectOrganizations  : get organization by tag
// Returns : information organization(JSON)
func (ctr *OrgController) GetOrganizations(c echo.Context) error {
	org, err := ctr.OrgRepo.SelectOrganizations()

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Organization is not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var dataResponse []map[string]interface{}
	for _, record := range org {
		data := map[string]interface{}{
			"id":           record.ID,
			"name":         record.Name,
			"tag":          record.Tag,
			"description":  record.Description,
			"phone_number": record.PhoneNumber,
		}

		dataResponse = append(dataResponse, data)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    dataResponse,
	})
}

// CheckOrganization : call model checkOrganization() to check organization exist
// Params            :
// Returns           : return data with struct JsonResponse
func (ctr *OrgController) CheckOrganization(c echo.Context) error {

	checkOrganizationParams := new(param.CheckOrganizationParams)
	if err := c.Bind(checkOrganizationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	org, err := ctr.OrgRepo.FindOrganizationByTag(checkOrganizationParams.OrganizationTag)
	if err == nil && org.Tag != "" {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Organization already registered.",
		})
	}

	if err != nil && err.Error() == pg.ErrNoRows.Error() {
		returnObj := map[string]interface{}{
			"tag":  checkOrganizationParams.OrganizationTag,
			"name": checkOrganizationParams.OrganizationName,
		}

		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.SuccessResponseCode,
			Message: "Organization can be register.",
			Data:    returnObj,
		})
	}

	return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
		Status:  cf.FailResponseCode,
		Message: "System error",
	})
}

// RegisterOrganization : call model registerOrganization() to insert organization
// Params            :
// Returns           : return data with struct JsonResponse
func (ctr *OrgController) RegisterOrganization(c echo.Context) error {

	registerOrganizationParams := new(param.RegisterOrganizationParams)
	if err := c.Bind(registerOrganizationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	regOrg, err := ctr.RegCodeRepo.GetRegCode(registerOrganizationParams.Code)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Your registration link is wrong.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if regOrg.ExpiredAt.Before(utils.TimeNowUTC()) {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Code Expired . Please register new code",
		})
	}

	// -------------------------------------> process save organization <------------------------------------------------------
	registerOrganizationParams.GoogleID = regOrg.GoogleID
	organization, email, userID, err := ctr.OrgRepo.SaveOrganization(ctr.UserRepo, ctr.RegCodeRepo, registerOrganizationParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Organization register fail",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// -----------------------------------> Auto Insert functions for user with status = ACCESSFUNC <-------------------------------
	var orgModule []m.OrgModule
	modules := []byte(cf.ModulesOrg)

	json.Unmarshal(modules, &orgModule)

	dataInit := &param.DataInitOrg{
		OrganizationId: organization.ID,
		Modules:        orgModule,
	}
	err = ctr.AdminRepo.InsertOrgModule(dataInit)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// -----------------------------------> Auto Insert functions for user with status = ACCESSFUNC <-------------------------------
	functionsOrg, err := ctr.AdminRepo.SelectFunctionsOrg(
		organization.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	for _, function := range functionsOrg {
		err = ctr.AdminRepo.InsertUserPermission(
			organization.ID,
			userID,
			function.Id,
			cf.ACCESSFUNC,
		)
	}

	ctr.InitSmtp(os.Getenv("MAIL_ADDRESS"), os.Getenv("MAIL_PASSWORD"))

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{email}
	sampleData.OrgTag = organization.Tag
	if err := ctr.SendMail("Micro Erp Successful Register", sampleData, cf.RegisterSuccessfulTemplate); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	dataResponse := map[string]interface{}{
		"id":   organization.ID,
		"tag":  organization.Name,
		"name": organization.Tag,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Organization register successfull.",
		Data:    dataResponse,
	})
}

// RegisterInviteLink : call model registerOrganization() to insert organization
// Params             :
// Returns            : return data with struct JsonResponse
func (ctr *OrgController) RegisterInviteLink(c echo.Context) error {

	registerInviteLinkParams := new(param.RegisterInviteLinkParams)

	if err := c.Bind(registerInviteLinkParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	// -------------------------------------> process register user to organization <------------------------------------------
	requestObj, userID, err := ctr.OrgRepo.SaveInviteRegister(ctr.UserRepo, ctr.RegCodeRepo, ctr.RequestRepo, registerInviteLinkParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Register fail",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	orgObj, err := ctr.OrgRepo.GetOrganizationByID(requestObj.OrganizationID)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Register fail",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	ctr.InitSmtp(orgObj.Email, orgObj.EmailPassword)

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{requestObj.Email}
	sampleData.OrgTag = orgObj.Tag
	if err := ctr.SendMail("Micro Erp Successful Register", sampleData, cf.RegisterSuccessfulTemplate); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// -----------------------------------> Auto Insert functions for user with status = NOTACCESSFUNC <-------------------------------
	functionsOrg, err := ctr.AdminRepo.SelectFunctionsOrg(
		orgObj.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	for _, function := range functionsOrg {
		err = ctr.AdminRepo.InsertUserPermission(
			orgObj.ID,
			userID,
			function.Id,
			cf.ACCESSFUNC,
		)
	}

	dataResponse := map[string]interface{}{
		"id":   orgObj.ID,
		"tag":  orgObj.Name,
		"name": orgObj.Tag,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Register successful.",
		Data:    dataResponse,
	})
}

func (ctr *OrgController) EditOrganizationEmail(c echo.Context) error {
	emailForOrganizationParams := new(param.EmailForOrganizationParams)
	if err := c.Bind(emailForOrganizationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(emailForOrganizationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if err := ctr.OrgRepo.UpdateEmailForOrganization(
		userProfile.OrganizationID,
		os.Getenv("MAIL_ADDRESS"),
		os.Getenv("MAIL_PASSWORD"),
		emailForOrganizationParams,
		userProfile.Organization.SettingStep,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update email for organization successful. Please check email test",
	})
}

func (ctr *OrgController) GetOrganizationSetting(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	dataResponse := map[string]interface{}{
		"email":                    userProfile.Organization.Email,
		"password":                 userProfile.Organization.EmailPassword,
		"expiration_reset_day_off": userProfile.Organization.ExpirationResetDayOff,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get setting for organization successful.",
		Data:    dataResponse,
	})
}

func (ctr *OrgController) EditExpirationResetDayOff(c echo.Context) error {
	params := new(param.EditExpirationResetDayOffParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if err := ctr.OrgRepo.UpdateExpirationResetDayOff(userProfile.OrganizationID, params.Expiration); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit expiration reset day off successful",
	})
}
