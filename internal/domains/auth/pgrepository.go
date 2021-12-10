package auth

import (
	b64 "encoding/base64"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// PgRegCodeRepository comment
type PgRegCodeRepository struct {
	cm.AppRepository
}

// NewPgRegCodeRepository comment
func NewPgRegCodeRepository(logger echo.Logger) (repo *PgRegCodeRepository) {
	repo = &PgRegCodeRepository{}
	repo.Init(logger)
	return
}

// CheckRegCodeByEmail : check email exist and last send within 15 minutes
// Params         : requestEmail - input email
// Returns        : true, false - email valid or sendable
func (repo *PgRegCodeRepository) CheckRegCodeByEmail(requestEmail string) (bool, error) {
	registrationCode := m.RegistrationCode{}
	err := repo.DB.Model(&registrationCode).
		Column("created_at").
		Order("created_at DESC").
		Limit(1).
		Where("email = ?", requestEmail).
		Select()

	if err != nil {
		repo.Logger.Error(err)
		if err.Error() == pg.ErrNoRows.Error() {
			return true, nil
		}

		return false, err
	}

	// add 15 minutes to time
	lastSend := registrationCode.CreatedAt.Add(15 * time.Minute)
	if lastSend.After(utils.TimeNowUTC()) {
		return false, nil
	}

	return true, nil
}

// InsertNewRegCode : generate register code and insert to db
// Params           : requestEmail - input email
// Returns          : true, false - insert result
func (repo *PgRegCodeRepository) InsertNewRegCode(createRegRequestParams *param.CreateRegRequestParams) (string, bool, error) {
	requestEmail := createRegRequestParams.RequestEmail
	registrationCode := m.RegistrationCode{}
	isSendable, err := repo.CheckRegCodeByEmail(requestEmail)

	if err != nil {
		repo.Logger.Error(err)
		return registrationCode.Code, false, err
	}

	if !isSendable && err == nil {
		return registrationCode.Code, false, nil
	}

	if isSendable {

		expiredTime, encodeString := InitRegCodeData(utils.TimeNowUTC(), requestEmail)
		registrationCode = m.RegistrationCode{
			Email:     requestEmail,
			Code:      encodeString,
			ExpiredAt: expiredTime,
			GoogleID:  createRegRequestParams.GoogleID,
		}
		err = repo.DB.Insert(&registrationCode)

		if err != nil {
			repo.Logger.Error(err)
			return registrationCode.Code, isSendable, err
		}

		isSendable = true
	}

	return registrationCode.Code, isSendable, err
}

// GetRegCode : check valid registration code
// Params     : registrationCode - registration code
// Returns    : return object of record that belong to registrationCode
func (repo *PgRegCodeRepository) GetRegCode(registrationCode string) (m.RegistrationCode, error) {
	registration := m.RegistrationCode{}
	err := repo.DB.Model(&registration).
		Limit(1).
		Where("code = ?", registrationCode).
		Select()

	return registration, err
}

// UpdateExpiredDateTx : update expired_date to registration_codes for use in transaction
// Params            : code, organizationTag, email, firstname, lastname, password - input data to insert
// Returns           : error
func (repo *PgRegCodeRepository) UpdateExpiredDateTx(tx *pg.Tx, code string) error {
	_, err := tx.Model(&m.RegistrationCode{ExpiredAt: utils.TimeNowUTC()}).
		Column("expired_at", "updated_at").
		Where("code = ?", code).
		Update()

	return err
}

// InitRegCodeData : generate register code and insert to db
// Params          : currentTime , requestEmail - input email
// Returns         : expiredTime , encodeString - output encode string
func InitRegCodeData(currentTime time.Time, requestEmail string) (time.Time, string) {

	expiredTime := currentTime.Add(cf.ExpiredHours * time.Hour)
	codeString := requestEmail + currentTime.Format(cf.FormatDate)
	// encode base64
	encodeString := b64.StdEncoding.EncodeToString([]byte(codeString))
	// >> decode : sDec, _ := b64.StdEncoding.DecodeString(sEnc)

	return expiredTime, encodeString
}

// InsertRegCodeWithTx : insert data to Registration code
// Params              : email - input email
// Returns             : RegistrationCode object , error
func (repo *PgRegCodeRepository) InsertRegCodeWithTx(tx *pg.Tx, email string, regRequestID int) (m.RegistrationCode, error) {
	expiredTime, encodeString := InitRegCodeData(utils.TimeNowUTC(), email)

	registrationCode := m.RegistrationCode{
		Email:                 email,
		Code:                  encodeString,
		RegistrationRequestID: regRequestID,
		ExpiredAt:             expiredTime,
	}

	err := tx.Insert(&registrationCode)

	if err != nil {
		repo.Logger.Error(err)
		return registrationCode, err
	}

	return registrationCode, err
}

// GetNewRegistCodeByRequestID : get newest registration code by request id
// Params         : requestEmail - input email
// Returns        : true, false - email valid or sendable
func (repo *PgRegCodeRepository) GetNewRegistCodeByRequestID(requestID int) (m.RegistrationCode, error) {
	registration := m.RegistrationCode{}
	err := repo.DB.Model(&registration).
		Where("registration_request_id = ?", requestID).
		Order("created_at DESC").
		First()
	return registration, err
}

// UpdateGoogleID : update google id
func (repo *PgRegCodeRepository) UpdateGoogleID(updateGoogleUser *param.UpdateGoogleUser) error {
	_, err := repo.DB.Model(&m.User{GoogleID: updateGoogleUser.GoogleID}).
		Column("google_id", "updated_at").
		Where("id = ?", updateGoogleUser.UserID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
