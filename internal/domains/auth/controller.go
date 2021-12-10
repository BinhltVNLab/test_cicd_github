package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"

	valid "github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg/v9"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthController struct {
	cm.BaseController
	email.SMTPGoMail

	RegCodeRepo rp.RegCodeRepository
	UserRepo    rp.UserRepository
}

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Load godotenv error : " + err.Error())
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "",
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func NewAuthController(logger echo.Logger, regCodeRepo rp.RegCodeRepository, userRepo rp.UserRepository) (ctr *AuthController) {
	ctr = &AuthController{cm.BaseController{}, email.SMTPGoMail{}, regCodeRepo, userRepo}
	ctr.Init(logger)
	ctr.InitSmtp(os.Getenv("MAIL_ADDRESS"), os.Getenv("MAIL_PASSWORD"))
	return
}

// createTokenLogin : create token login
// Params  : user id
// Returns : string token login
func createTokenLogin(userID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = userID
	claims["exp"] = utils.TimeNowUTC().Add(time.Hour * 72).Unix()

	keyTokenAuth := utils.GetKeyToken()
	t, err := token.SignedString([]byte(keyTokenAuth))

	return t, err
}

// Login   : check information Login (email, password, organization_id) if valid regsiter and send Token to SPA
// Params  : echo.Context
// Returns : Token Login (JSON)
func (ctr *AuthController) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := utils.GetSHA256Hash(c.FormValue("password"))
	organizationID, err := strconv.Atoi(c.FormValue("organization_id"))

	if !valid.IsEmail(email) {
		return c.JSON(http.StatusUnauthorized, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid email",
		})
	}

	if err != nil {
		return c.JSON(http.StatusUnauthorized, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid organization",
		})
	}

	// get ID user login in DB
	idUserLogin, err := ctr.UserRepo.GetLoginUserID(email, password, organizationID)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			//select no rows in database
			return c.JSON(http.StatusUnauthorized, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not exist or password wrong",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	err = ctr.UserRepo.UpdateLastLogin(idUserLogin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error ",
		})
	}

	tokenLogin, err := createTokenLogin(idUserLogin)

	objToken := map[string]string{
		"token": tokenLogin,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    objToken,
	})
}

// Logout  : Destroy Token login
// Params  : echo.Context
// Returns : JSON
func (ctr *AuthController) Logout(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	claims["exp"] = utils.TimeNowUTC()

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
	})
}

// CreateRegRequest : call model insertNewEmailRegister() to check email exist
// Params     :
// Returns    : return data with struct JsonResponse
func (ctr *AuthController) CreateRegRequest(c echo.Context) error {

	createRegRequestParams := new(param.CreateRegRequestParams)
	if err := c.Bind(createRegRequestParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	code, result, err := ctr.RegCodeRepo.InsertNewRegCode(createRegRequestParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	returnObj := map[string]string{
		"email": createRegRequestParams.RequestEmail,
	}

	if !result {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Within 15 minutes since last times send mail",
			Data:    returnObj,
		})
	}

	sampleData := new(param.SampleData)
	sampleData.SendTo = []string{createRegRequestParams.RequestEmail}
	sampleData.URL = os.Getenv("BASE_SPA_URL") + "/organization/create-organization/" + code

	if err := ctr.SendMail("Micro Erp Registration Email", sampleData, cf.CreateOrganizationTemplate); err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "",
		Data:    returnObj,
	})
}

// CheckRegistrationCode : call model checkCode() to check code valid
// Params                :
// Returns               : return data with struct JsonResponse
func (ctr *AuthController) CheckRegistrationCode(c echo.Context) error {

	checkRegistrationCodeParams := new(param.CheckRegistrationCodeParams)
	if err := c.Bind(checkRegistrationCodeParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	result, err := ctr.RegCodeRepo.GetRegCode(checkRegistrationCodeParams.Code)

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

	if result.ExpiredAt.Before(utils.TimeNowUTC()) {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Code expired. Please register new code.",
		})
	}

	returnObj := map[string]interface{}{
		"request_id": result.RegistrationRequestID,
		"email":      result.Email,
		"code":       result.Code,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "OK we can process register organization.",
		Data:    returnObj,
	})
}

// OauthGoogleLogin : use public key and secret key to create link accesss page login google
// when login google success go to page "login google callback"
// Params  : echo.Context
// Returns : redirect to link google api
func (ctr *AuthController) OauthGoogleLogin(c echo.Context) error {
	organizationID := c.FormValue("organization_id")
	// create unique id and use it to make state
	state := utils.GetUniqueString()
	var expiration = utils.TimeNowUTC().Add(time.Minute * 3)

	// create cookie organization
	cookieOrganizationID := new(http.Cookie)
	cookieOrganizationID.Name = "organization_id"
	cookieOrganizationID.Value = organizationID
	cookieOrganizationID.Expires = expiration
	c.SetCookie(cookieOrganizationID)

	// create cookie login state
	cookieOauthState := new(http.Cookie)
	cookieOauthState.Name = "oauthstate"
	cookieOauthState.Value = state
	cookieOauthState.Expires = expiration
	c.SetCookie(cookieOauthState)

	// set RedirectURL for login google
	googleOauthConfig.RedirectURL = os.Getenv("BASE_API_URL") + "/auth/login-google-callback"
	url := googleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
	return nil
}

// OauthGoogleCallback : call back when login google
// Params  : echo.Context
// Returns : Token Login (JSON)
func (ctr *AuthController) OauthGoogleCallback(c echo.Context) error {
	state := c.FormValue("state")
	code := c.FormValue("code")
	// link callback to spa
	urlSpa := os.Getenv("BASE_SPA_URL") + "/user/login?type=social"

	organizationIDCookie, err := c.Cookie("organization_id")

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed organization"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
	}

	organizationID, err := strconv.Atoi(organizationIDCookie.Value)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Invalid organization"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
	}

	urlSpa = urlSpa + "&organization_id=" + organizationIDCookie.Value

	oauthStateCookie, err := c.Cookie("oauthstate")

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed oauthstate"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if state != oauthStateCookie.Value {
		linkRedirect := urlSpa + "&error_message=" + "Invalid oauth state"
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Code exchange failed"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed getting user info"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed read post"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	googleProfile := new(param.GoogleProfile)
	err = json.Unmarshal(contents, &googleProfile)
	utils.PrintVars(os.Stdout, true, contents)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "JSON syntax error"
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if !googleProfile.VerifiedEmail {
		linkRedirect := urlSpa + "&error_message=login by google error"
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	user, err := ctr.UserRepo.GetUserByEmailOrganizationID(googleProfile.Email, organizationID)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			linkRedirect := urlSpa + "&error_message=" + "User account does not exist"
			c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
			return nil
		}

		linkRedirect := urlSpa + "&error_message=" + "System error"
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	err = ctr.UserRepo.UpdateLastLogin(user.ID)
	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "System error"
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if user.GoogleID == "" {
		googleUser := new(param.UpdateGoogleUser)
		googleUser.UserID = user.ID
		googleUser.GoogleID = googleProfile.GoogleID

		err = ctr.RegCodeRepo.UpdateGoogleID(googleUser)

		if err != nil {
			linkRedirect := urlSpa + "&error_message=" + "System error"
			c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
			return nil
		}
	}

	tokenLogin, err := createTokenLogin(user.ID)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "System error"
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	linkRedirect := urlSpa + "&token=" + tokenLogin
	c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
	return nil
}

// RegisterOrgGoogle : Register org with google
// Params  : echo.Context
// Returns : Redirect to link google api
func (ctr *AuthController) RegisterOrgGoogle(c echo.Context) error {
	// create unique id and use it to make state
	state := utils.GetUniqueString()
	var expiration = utils.TimeNowUTC().Add(time.Minute * 3)

	// create cookie login state
	cookieOauthState := new(http.Cookie)
	cookieOauthState.Name = "oauthstate"
	cookieOauthState.Value = state
	cookieOauthState.Expires = expiration
	c.SetCookie(cookieOauthState)

	googleOauthConfig.RedirectURL = os.Getenv("BASE_API_URL") + "/registration/register-org-google-callback"
	url := googleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
	return nil
}

// RegisterOrgGoogleCallBack : Register org with google
// Params  : echo.Context
// Returns : Redirect to SPA create organization with request
func (ctr *AuthController) RegisterOrgGoogleCallBack(c echo.Context) error {
	state := c.FormValue("state")
	code := c.FormValue("code")
	// link callback to spa
	urlSpa := os.Getenv("BASE_SPA_URL") + "/registration?type=social"

	oauthStateCookie, err := c.Cookie("oauthstate")

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed oauthstate."
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if state != oauthStateCookie.Value {
		linkRedirect := urlSpa + "&error_message=" + "Invalid oauth state."
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Code exchange failed."
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed getting user info."
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "Failed read post."
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	googleProfile := new(param.GoogleProfile)
	err = json.Unmarshal(contents, &googleProfile)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=" + "JSON syntax error."
		ctr.Logger.Error(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if !googleProfile.VerifiedEmail {
		linkRedirect := urlSpa + "&error_message=Sign up by google error."
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	createRegRequestParams := &param.CreateRegRequestParams{
		RequestEmail: googleProfile.Email,
		GoogleID:     googleProfile.GoogleID,
	}

	codeRegister, limitSend, err := ctr.RegCodeRepo.InsertNewRegCode(createRegRequestParams)

	if err != nil {
		linkRedirect := urlSpa + "&error_message=System error."
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	if !limitSend {
		linkRedirect := urlSpa + "&error_message=Within 15 minutes since last times send mail."
		c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
		return nil
	}

	linkRedirect := urlSpa + "&reg_code=" + codeRegister
	c.Redirect(http.StatusTemporaryRedirect, linkRedirect)
	return nil
}

func (ctr *AuthController) DownloadTemplate(c echo.Context) error {
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
		if userProfile.LanguageId == cf.EnLanguageId || userProfile.LanguageId == cf.VnLanguageId {
			return c.Attachment("internal/platform/excel/template/import-email.xlsx", "import-email.xlsx")
		}

		return c.Attachment("internal/platform/excel/template/import-email-jp.xlsx", "import-email-jp.xlsx")
	}

	if userProfile.LanguageId == cf.EnLanguageId || userProfile.LanguageId == cf.VnLanguageId {
		return c.Attachment("internal/platform/excel/template/import-email.csv", "import-email.csv")
	}

	return c.Attachment("internal/platform/excel/template/import-email-jp.csv", "import-email-jp.csv")
}
