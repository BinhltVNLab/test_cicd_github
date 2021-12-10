package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// RegCodeRepository interface
type RegCodeRepository interface {
	CheckRegCodeByEmail(requestEmail string) (bool, error)
	InsertNewRegCode(createRegRequestParams *param.CreateRegRequestParams) (string, bool, error)
	InsertRegCodeWithTx(tx *pg.Tx, emailAddr string, organizationID int) (m.RegistrationCode, error)
	GetRegCode(registrationCode string) (m.RegistrationCode, error)
	UpdateExpiredDateTx(tx *pg.Tx, code string) error
	GetNewRegistCodeByRequestID(requestID int) (m.RegistrationCode, error)
	UpdateGoogleID(googleUser *param.UpdateGoogleUser) error
}
