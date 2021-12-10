package organization

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/email"
	"os"
	"strings"
)

type PgOrgRepository struct {
	cm.AppRepository
	email.SMTPGoMail
}

func NewPgOrgRepository(logger echo.Logger) (repo *PgOrgRepository) {
	repo = &PgOrgRepository{cm.AppRepository{}, email.SMTPGoMail{}}
	repo.Init(logger)
	return
}

// FindOrganizationByTag : get organization by tag
// Params              : tag - POST
// Returns             : information Organization(Object)
func (repo *PgOrgRepository) FindOrganizationByTag(tag string) (m.Organization, error) {
	objOrg := m.Organization{}
	err := repo.DB.Model(&objOrg).
		Limit(1).
		Column("id", "name", "tag").
		Where("tag = ?", strings.ToUpper(tag)).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return objOrg, err
}

// SelectOrganizations : get organization by tag
// Params              : tag - POST
// Returns             : information Organization(Object)
func (repo *PgOrgRepository) SelectOrganizations() ([]m.Organization, error) {
	objOrg := []m.Organization{}
	err := repo.DB.Model(&objOrg).
		Column("org.id", "org.name", "org.tag", "org.description", "org.phone_number").
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}
 
	return objOrg, err
}

// GetOrganizationByID : get organization by ID
// Params              : id
// Returns             : information Organization(Object)
func (repo *PgOrgRepository) GetOrganizationByID(id int) (m.Organization, error) {
	objOrg := m.Organization{}
	err := repo.DB.Model(&objOrg).
		Limit(1).
		Column("id", "name", "tag", "email", "email_password").
		Where("id = ?", id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return objOrg, err
}

// SaveOrganization : insert data to organizations, users, user_profiles
// Params               : organizationName, organizationTag, email, firstname, lastname, password - input data to insert
// Returns              : return object of record organizations just inserted
func (repo *PgOrgRepository) SaveOrganization(userRepo rp.UserRepository, regCodeRepo rp.RegCodeRepository, registerOrganizationParams *param.RegisterOrganizationParams) (m.Organization, string, int, error) {
	organization := m.Organization{}

	var userID int
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		organization, transErr = repo.InsertOrganizationWithTx(tx, registerOrganizationParams.OrganizationName, registerOrganizationParams.OrganizationTag)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		registerUser := param.RegisterUserParams{
			OrganizationID: organization.ID,
			Email:          registerOrganizationParams.Email,
			Password:       registerOrganizationParams.Password,
			RoleID:         cf.GeneralManagerRoleID,
			GoogleID:       registerOrganizationParams.GoogleID,
			LanguageId:     cf.EnLanguageId,
		}
		user, transErr := userRepo.InsertUserWithTx(tx, registerUser)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		userID = user.ID
		transErr = userRepo.InsertUserProfileWithTx(tx, user.ID, registerOrganizationParams.FirstName, registerOrganizationParams.LastName)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		transErr = regCodeRepo.UpdateExpiredDateTx(tx, registerOrganizationParams.Code)

		return transErr
	})

	return organization, registerOrganizationParams.Email, userID, err
}

// InsertOrganizationWithTx : insert data to organizations
// Params             : organizationName, organizationTag, ecurrentTime - input data to insert
// Returns            : return organizations object , error
func (repo *PgOrgRepository) InsertOrganizationWithTx(tx *pg.Tx, organizationName string, organizationTag string) (m.Organization, error) {
	organization := m.Organization{
		Name:        organizationName,
		Tag:         strings.ToUpper(organizationTag),
		PhoneNumber: "",
		Address:     "",
		Description: "",
	}
	err := tx.Insert(&organization)
	return organization, err
}

// SaveInviteRegister : insert data to organizations, users, user_profiles
// Params               : organizationName, organizationTag, email, firstname, lastname, password - input data to insert
// Returns              : return object of record organizations just inserted
func (repo *PgOrgRepository) SaveInviteRegister(userRepo rp.UserRepository, regCodeRepo rp.RegCodeRepository, requestRepo rp.RegistRequestRepository,
	registerInviteLinkParams *param.RegisterInviteLinkParams) (m.RegistrationRequest, int, error) {
	requestObj := m.RegistrationRequest{}
	var userID int
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		requestObj, transErr = requestRepo.GetRegRequestsByID(registerInviteLinkParams.RequestID)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		registerUser := param.RegisterUserParams{
			OrganizationID: requestObj.OrganizationID,
			Email:          registerInviteLinkParams.Email,
			Password:       registerInviteLinkParams.Password,
			RoleID:         cf.UserRoleID,
			LanguageId:     cf.EnLanguageId,
		}

		user, transErr := userRepo.InsertUserWithTx(tx, registerUser)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		userID = user.ID
		transErr = userRepo.InsertUserProfileWithTx(tx, userID, registerInviteLinkParams.FirstName, registerInviteLinkParams.LastName)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		transErr = regCodeRepo.UpdateExpiredDateTx(tx, registerInviteLinkParams.Code)

		transErr = requestRepo.UpdateStatusRequestWithTX(tx, registerInviteLinkParams.RequestID, cf.RegisteredRequestStatus)

		return transErr
	})

	return requestObj, userID, err
}

func (repo *PgOrgRepository) UpdateEmailForOrganization(
	organizationId int,
	emailAddress string,
	emailPass string,
	emailForOrganizationParams *param.EmailForOrganizationParams,
	settingStep int,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		organization := m.Organization{
			Email:         emailAddress,
			EmailPassword: emailPass,
		}

		if settingStep == 0 {
			organization.SettingStep = cf.ORGANIZATIONEMAILSETTING
		}

		q := tx.Model(&organization).Column("email", "email_password", "updated_at")
		if settingStep == 0 {
			q.Column("setting_step")
		}
		_, transErr = q.Where("id = ?", organizationId).Update()

		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		repo.InitSmtp(emailAddress, emailPass)
		sampleData := new(param.SampleData)
		sampleData.SendTo = []string{emailForOrganizationParams.EmailTest}
		sampleData.URL = os.Getenv("BASE_SPA_URL") + "/settings/branch"
		if transErr := repo.SendMail("Test Send Email", sampleData, cf.SendTestMailTemplate); transErr != nil {
			repo.Logger.Error(transErr)
			organizationObj, err := repo.SelectEmailAndPassword(organizationId)
			if err != nil && err.Error() != pg.ErrNoRows.Error() {
				repo.Logger.Error(err)
				return transErr
			}
			repo.InitSmtp(organizationObj.Email, organizationObj.EmailPassword)
			return transErr
		}

		return transErr
	})

	return err
}

func (repo *PgOrgRepository) SelectEmailAndPassword(Id int) (m.Organization, error) {
	organization := m.Organization{}
	err := repo.DB.Model(&organization).
		Column("email", "email_password").
		Where("id = ?", Id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return organization, err
}

func (repo *PgOrgRepository) UpdateSettingStepWithTx(tx *pg.Tx, Id int, step int) error {
	_, err := tx.Model(&m.Organization{SettingStep: step}).
		Column("setting_step", "updated_at").
		Where("id = ?", Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgOrgRepository) UpdateExpirationResetDayOff(id int, expiration int) error {
	_, err := repo.DB.Model(&m.Organization{ExpirationResetDayOff: expiration}).
		Column("expiration_reset_day_off", "updated_at").
		Where("id = ?", id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
