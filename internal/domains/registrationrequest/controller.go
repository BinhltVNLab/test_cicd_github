package registrationrequest

import (
	"net/http"
	"os"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// RegistRequestController struct implement OrgRepository, RegCodeRepository
type RegistRequestController struct {
	cm.BaseController
	email.SMTPGoMail

	UserRepo          rp.UserRepository
	RegCodeRepo       rp.RegCodeRepository
	OrgRepo           rp.OrgRepository
	RegistRequestRepo rp.RegistRequestRepository
}

// NewRegistRequestController import regRepo, orgRepo
func NewRegistRequestController(logger echo.Logger, userRepo rp.UserRepository, regCodeRepo rp.RegCodeRepository, orgRepo rp.OrgRepository, registRequestRepo rp.RegistRequestRepository) (ctr *RegistRequestController) {
	ctr = &RegistRequestController{cm.BaseController{}, email.SMTPGoMail{}, userRepo, regCodeRepo, orgRepo, registRequestRepo}
	ctr.Init(logger)
	return
}

// RequestRegistration : call model RequestRegistration() to insert organization
// Params              : form value : organizationID, email , message
// Returns             : return data with struct JsonResponse
func (ctr *RegistRequestController) RequestRegistration(c echo.Context) error {

	requestRegistrationParams := new(param.RequestRegistrationParams)

	if err := c.Bind(requestRegistrationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid params",
			Data:    err,
		})
	}

	registrationRequestOjb, err := ctr.RegistRequestRepo.GetRegRequests(requestRegistrationParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// check email is in use or not
	errMail, err := ctr.RegistRequestRepo.CheckExistRequestUser(
		ctr.UserRepo,
		[]string{requestRegistrationParams.EmailAddr},
		requestRegistrationParams.OrganizationID,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if len(errMail) > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Email is requested or registered",
			Data:    errMail,
		})
	}

	org, err := ctr.OrgRepo.SelectEmailAndPassword(requestRegistrationParams.OrganizationID)
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

	// no record in db
	if registrationRequestOjb.ID == 0 {
		// start insert new request
		registrationRequestOjb, err := ctr.RegistRequestRepo.InsertRegRequest(cf.UserRequestType, cf.PendingRequestStatus, requestRegistrationParams)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		ctr.InitSmtp(org.Email, org.EmailPassword)

		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{requestRegistrationParams.EmailAddr}

		// after insert success - send notice email
		if err := ctr.SendMail("Micro_Erp Success Request", sampleData, cf.SuccessRequestTemplate); err != nil {
			ctr.Logger.Error(err)
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.SuccessResponseCode,
			Message: "Your request has been sent. We will sent you an email about your request soon",
			Data:    registrationRequestOjb,
		})
	}

	// record exist in db
	if registrationRequestOjb.Email != "" {
		returnMessage := ""
		if registrationRequestOjb.Status == cf.DenyRequestStatus {
			returnMessage = "Your request has been denied"
		} else {
			returnMessage = "Your request have been sent before. Please wait for us to check"
		}
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: returnMessage,
		})
	}

	return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
		Status:  cf.FailResponseCode,
		Message: "System error",
	})
}

// InviteUser : call model inviteUser() to check email exist
// Params     : form value : organizationID, listEmail
// Returns    : return data with struct JsonResponse
func (ctr *RegistRequestController) InviteUser(c echo.Context) error {
	inviteUserParams := new(param.InviteUserParams)

	if err := c.Bind(inviteUserParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid params",
			Data:    err,
		})
	}

	// -------------------------------------> check valid email <--------------------------------------------------------------

	for _, emailAddr := range inviteUserParams.EmailAddr {
		if !valid.IsEmail(emailAddr) {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid email",
				Data:    emailAddr,
			})
		}
	}

	// -------------------------------------> get user info with Token <-------------------------------------------------------

	userProfile := c.Get("user_profile").(m.User)
	// -------------------------------------> check exist request and registerd user <-----------------------------------------

	errMail, err := ctr.RegistRequestRepo.CheckExistRequestUser(ctr.UserRepo, inviteUserParams.EmailAddr, userProfile.OrganizationID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if len(errMail) > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Email requested or registered",
			Data:    errMail,
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
	ctr.InitSmtp(org.Email, org.EmailPassword)

	// -------------------------------------> process insert to db <-----------------------------------------------------------

	regRequestIds, regCodeIds, sendMailList, err := ctr.RegistRequestRepo.SaveAcceptedRegRequest(
		ctr.RegCodeRepo,
		inviteUserParams.EmailAddr,
		userProfile.OrganizationID,
		"",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var errorEmails []string
	for mail, code := range sendMailList {
		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{mail}
		sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/create-organization/" + code

		if err := ctr.SendMail("Micro Erp Registration Email", sampleData, cf.CreateOrganizationTemplate); err != nil {
			ctr.Logger.Error(err, " - error mail : ", err)
			errorEmails = append(errorEmails, mail)
			removeErr := ctr.RegistRequestRepo.DeleteInviteRequests(regRequestIds, regCodeIds)
			if removeErr != nil {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System error",
				})
			}
		}
	}

	if len(errorEmails) > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Send email fail",
			Data:    errorEmails,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Invite email was sent",
	})
}

// SearchListRequest : search, filter, order registration request
// Params            : echo.Context
// Returns           : return data with struct JsonResponse
func (ctr *RegistRequestController) SearchListRequest(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	manageRequestParams := new(param.ManageRequestParams)
	manageRequestParams.OrganizationID = userProfile.OrganizationID
	manageRequestParams.RowPerPage = 10

	if err := c.Bind(manageRequestParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	// check field type date have format format 2006/01/02 (YYYY/mm/dd)
	valid.TagMap["formatDisplayDate"] = utils.ValidatorFormatDisplayDate()

	_, err := valid.ValidateStruct(manageRequestParams)

	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	// return error if date from after date to
	if manageRequestParams.DateFrom.IsZero() && manageRequestParams.DateTo.IsZero() {
		if manageRequestParams.DateFrom.After(manageRequestParams.DateTo) {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Date from should be less than or equal date to",
			})
		}
	}

	listRequest, totalRow, err := ctr.RegistRequestRepo.GetListRequest(manageRequestParams)

	if err != nil {
		if err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	listRequestResponse := []map[string]interface{}{}
	lengListRequest := len(listRequest)

	for i := 0; i < lengListRequest; i++ {
		itemListRequest := listRequest[i]
		typeName := utils.GetNameTypeRegistRequests(itemListRequest.Type)
		statusName := utils.GetNameStatusRegistRequests(itemListRequest.Status)
		checkAllowResend := false
		newestRegistCode, err := ctr.RegCodeRepo.GetNewRegistCodeByRequestID(itemListRequest.ID)

		if err != nil {
			if err.Error() != pg.ErrNoRows.Error() {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System error",
				})
			}
		}

		if newestRegistCode.ExpiredAt.After(utils.TimeNowUTC()) {
			checkAllowResend = true
		}

		itemDataResponse := map[string]interface{}{
			"id":           itemListRequest.ID,
			"email":        itemListRequest.Email,
			"type":         itemListRequest.Type,
			"type_name":    typeName,
			"status":       itemListRequest.Status,
			"status_name":  statusName,
			"message":      itemListRequest.Message,
			"time_request": itemListRequest.CreatedAt.Format(cf.FormatTimeDisplay),
			"allow_resend": checkAllowResend,
		}

		listRequestResponse = append(listRequestResponse, itemDataResponse)
	}

	pagination := map[string]interface{}{
		"current_page": manageRequestParams.CurrentPage,
		"total_row":    totalRow,
		"row_perpage":  manageRequestParams.RowPerPage,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination":   pagination,
			"list_request": listRequestResponse,
		},
	})
}

// UpdateRequestStatus : accept or deny request
// Params              : echo.Context
// Returns             : return data with struct JsonResponse
func (ctr *RegistRequestController) UpdateRequestStatus(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	acceptOrDenyParam := new(param.UpdateRequestStatusParams)

	if err := c.Bind(acceptOrDenyParam); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	requestObj, err := ctr.RegistRequestRepo.GetRegRequestsByID(acceptOrDenyParam.RequestID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Request is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
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
	ctr.InitSmtp(org.Email, org.EmailPassword)

	acceptOrDenyParam.Email = requestObj.Email

	if acceptOrDenyParam.Status == cf.AcceptRequestStatus {
		regCode, err := ctr.RegistRequestRepo.AcceptRequest(ctr.RegCodeRepo, acceptOrDenyParam)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{acceptOrDenyParam.Email}
		sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/create-organization/" + regCode.Code
		if err := ctr.SendMail("Micro Erp Registration Email", sampleData, cf.CreateOrganizationTemplate); err != nil {
			ctr.Logger.Error(err)
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	} else if acceptOrDenyParam.Status == cf.DenyRequestStatus {
		err = ctr.RegistRequestRepo.DenyRequest(acceptOrDenyParam)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{acceptOrDenyParam.Email}
		sampleData.Content = "Your request join to " + userProfile.Organization.Name + " organization is deny. We sorry about that."

		err = ctr.SendMail("Announce from Micro Erp", sampleData, cf.TemplateSendMailAnnounce)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
	})
}

// ResendEmailRegister : accept or deny request
// Params              : echo.Context
// Returns             : return data with struct JsonResponse
func (ctr *RegistRequestController) ResendEmailRegister(c echo.Context) error {
	resendEmailParams := new(param.ResendEmailParams)

	if err := c.Bind(resendEmailParams); err != nil {
		msgErrBind := err.Error()
		fieldErr := utils.GetFieldBindForm(msgErrBind)

		if fieldErr != "" {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field " + fieldErr,
			})
		}
	}

	requestObj, err := ctr.RegistRequestRepo.GetRegRequestsByID(resendEmailParams.RequestID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Request is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	newestRegistCode, err := ctr.RegCodeRepo.GetNewRegistCodeByRequestID(resendEmailParams.RequestID)
	if err != nil {
		if err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	if !newestRegistCode.ExpiredAt.IsZero() && newestRegistCode.ExpiredAt.After(utils.TimeNowUTC()) {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The register code is not expired yet, please check your email and spam mail",
		})
	}

	resendEmailParams.Email = requestObj.Email
	regCodeObj, err := ctr.RegistRequestRepo.GetResendRegistCode(ctr.RegCodeRepo, resendEmailParams)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
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
	if org.Email == "" || org.EmailPassword == "" {
		return c.JSON(http.StatusUnprocessableEntity, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Your organization must have an email",
		})
	}

	ctr.InitSmtp(org.Email, org.EmailPassword)

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{regCodeObj.Email}
	sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/create-organization/" + regCodeObj.Code
	if err := ctr.SendMail("Micro Erp Registration Email", sampleData, cf.CreateOrganizationTemplate); err != nil {
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
