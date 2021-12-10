package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// OrgRepository interface
type OrgRepository interface {
	FindOrganizationByTag(tag string) (m.Organization, error)
	GetOrganizationByID(id int) (m.Organization, error)
	SaveOrganization(userRepo UserRepository, regCodeRepo RegCodeRepository, registerOrganizationParams *param.RegisterOrganizationParams) (m.Organization, string, int, error)
	SaveInviteRegister(userRepo UserRepository, regCodeRepo RegCodeRepository, requestRepo RegistRequestRepository, registerInviteLinkParams *param.RegisterInviteLinkParams) (m.RegistrationRequest, int, error)
	InsertOrganizationWithTx(tx *pg.Tx, organizationName string, organizationTag string) (m.Organization, error)
	UpdateEmailForOrganization(organizationId int, emailAddress string, emailPass string,
		emailForOrganizationParams *param.EmailForOrganizationParams, settingStep int) error
	SelectEmailAndPassword(Id int) (m.Organization, error)
	UpdateSettingStepWithTx(tx *pg.Tx, Id int, step int) error
	UpdateExpirationResetDayOff(id int, expiration int) error
	SelectOrganizations() ([]m.Organization, error)
}
