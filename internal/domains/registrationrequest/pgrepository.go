package registrationrequest

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// PgRequestRepository comment
type PgRequestRepository struct {
	cm.AppRepository
}

// NewPgRequestRepository comment
func NewPgRequestRepository(logger echo.Logger) (repo *PgRequestRepository) {
	repo = &PgRequestRepository{}
	repo.Init(logger)
	return
}

// GetRegRequests : check registration requests exist or not
// Params         : email, organizationID - input data to insert
// Returns        : RegistrationRequest object, error
func (repo *PgRequestRepository) GetRegRequests(requestRegistrationParams *param.RequestRegistrationParams) (m.RegistrationRequest, error) {

	registrationRequest := m.RegistrationRequest{}
	err := repo.DB.Model(&registrationRequest).
		Where("email = ?", requestRegistrationParams.EmailAddr).
		Where("organization_id = ?", requestRegistrationParams.OrganizationID).
		Select()

	return registrationRequest, err
}

// GetRegRequestsByID : check registration requests by id
// Params         : email, organizationID - input data to insert
// Returns        : RegistrationRequest object, error
func (repo *PgRequestRepository) GetRegRequestsByID(id int) (m.RegistrationRequest, error) {

	registrationRequest := m.RegistrationRequest{}
	err := repo.DB.Model(&registrationRequest).
		Where("id = ?", id).
		Select()

	return registrationRequest, err
}

// InsertRegRequest : insert new registration request data
// Params           : registration request info
// Returns          : RegistrationRequest object, error
func (repo *PgRequestRepository) InsertRegRequest(requestType int, requestStatus int, requestRegistrationParams *param.RequestRegistrationParams) (m.RegistrationRequest, error) {

	registrationRequest := m.RegistrationRequest{
		Type:           requestType,
		Status:         requestStatus,
		Email:          requestRegistrationParams.EmailAddr,
		OrganizationID: requestRegistrationParams.OrganizationID,
		Message:        requestRegistrationParams.Message,
	}

	err := repo.DB.Insert(&registrationRequest)

	repo.Logger.Error(err)
	return registrationRequest, err
}

// InsertRegRequestWithTx : insert new data to registration_requests
// Params                 : tx, registration request info
// Returns                : RegistrationRequest object, error
func (repo *PgRequestRepository) InsertRegRequestWithTx(tx *pg.Tx, requestType int, requestStatus int, requestEmail string, orgID int, message string) (m.RegistrationRequest, error) {

	registrationRequest := m.RegistrationRequest{
		Type:           requestType,
		Status:         requestStatus,
		Email:          requestEmail,
		OrganizationID: orgID,
		Message:        message,
	}

	err := tx.Insert(&registrationRequest)

	if err != nil {
		repo.Logger.Error(err)
		return registrationRequest, err
	}

	return registrationRequest, err
}

// SaveAcceptedRegRequest : generate register code and insert to db
// Params                 : regCodeRepo - RegCodeRepository import , registration request info
// Returns                : list email and code , error
func (repo *PgRequestRepository) SaveAcceptedRegRequest(
	regCodeRepo rp.RegCodeRepository,
	requestEmail []string,
	orgID int,
	message string,
) ([]int, []int, map[string]string, error) {
	var sendMailList = make(map[string]string)
	var regRequestIds, regCodeIds []int

	regCodeObj := m.RegistrationCode{}
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var err error
		for _, emailAddr := range requestEmail {
			regRequestObj, err := repo.InsertRegRequestWithTx(tx, cf.AdminInviteType, cf.AcceptRequestStatus, emailAddr, orgID, message)
			if err != nil {
				repo.Logger.Error(err)
				break
			}

			regCodeObj, err = regCodeRepo.InsertRegCodeWithTx(tx, emailAddr, regRequestObj.ID)
			if err != nil {
				repo.Logger.Error(err)
				break
			}

			regRequestIds = append(regRequestIds, regRequestObj.ID)
			regCodeIds = append(regCodeIds, regCodeObj.ID)
			sendMailList[regCodeObj.Email] = regCodeObj.Code
		}
		return err
	})

	return regRequestIds, regCodeIds, sendMailList, err
}

func (repo *PgRequestRepository) DeleteInviteRequests(regRequestIds []int, regCodeIds []int) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var tranErr error
		for _, regRequestId := range regRequestIds {
			_, tranErr := tx.Model(&m.RegistrationRequest{}).Where("id = ?", regRequestId).Delete()
			if tranErr != nil {
				repo.Logger.Error(tranErr)
				return tranErr
			}
		}

		for _, regCodeId := range regCodeIds {
			_, tranErr := tx.Model(&m.RegistrationCode{}).Where("id = ?", regCodeId).Delete()
			if tranErr != nil {
				repo.Logger.Error(tranErr)
				return tranErr
			}
		}

		return tranErr
	})

	return err
}

// UpdateStatusRequestWithTX : update status for request
// Params                : transaction object, request ID, status request
// Returns               : array list RegistrationRequest
func (repo *PgRequestRepository) UpdateStatusRequestWithTX(tx *pg.Tx, requestID int, status int) error {
	_, err := tx.Model(&m.RegistrationRequest{Status: status}).
		Column("status", "updated_at").
		Where("id = ?", requestID).
		Update()

	return err
}

// GetListRequest : get list registration request by organizationid, filter, pagination
// Params         : organizationID, filter data, pagination
// Returns        : list request
func (repo *PgRequestRepository) GetListRequest(manageRequestParams *param.ManageRequestParams) ([]m.RegistrationRequest, int, error) {
	registrationRequest := []m.RegistrationRequest{}
	totalRow, err := repo.DB.Model(&registrationRequest).
		Where("organization_id = ?", manageRequestParams.OrganizationID).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			queryCondition := q
			if manageRequestParams.Email != "" {
				queryCondition.Where("email LIKE ?", "%"+manageRequestParams.Email+"%")
			}

			if manageRequestParams.TypeRequest != 0 {
				queryCondition.Where("type = ?", manageRequestParams.TypeRequest)
			}

			if manageRequestParams.StatusRequest != 0 {
				queryCondition.Where("status = ?", manageRequestParams.StatusRequest)
			}

			if !manageRequestParams.DateFrom.IsZero() {
				queryCondition.Where("DATE(created_at) >= DATE(?)", manageRequestParams.DateFrom)
			}

			if !manageRequestParams.DateTo.IsZero() {
				queryCondition.Where("DATE(created_at) <= DATE(?)", manageRequestParams.DateTo)
			}

			return queryCondition, nil
		}).
		Offset((manageRequestParams.CurrentPage - 1) * manageRequestParams.RowPerPage).
		Order("created_at DESC").
		Limit(manageRequestParams.RowPerPage).
		SelectAndCount()

	return registrationRequest, totalRow, err
}

// AcceptRequest  : Update status accept request and create new code to register email
// Params         : regCodeRepo - RegCodeRepository import , param UpdateRequestStatusParams
// Returns        : register code, error
func (repo *PgRequestRepository) AcceptRequest(regCodeRepo rp.RegCodeRepository, updateRequestStatusParams *param.UpdateRequestStatusParams) (m.RegistrationCode, error) {
	regCode := m.RegistrationCode{}
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		err := repo.UpdateStatusRequestWithTX(tx, updateRequestStatusParams.RequestID, updateRequestStatusParams.Status)
		if err != nil {
			repo.Logger.Error(err)
			return err
		}

		regCode, err = regCodeRepo.InsertRegCodeWithTx(tx, updateRequestStatusParams.Email, updateRequestStatusParams.RequestID)
		if err != nil {
			repo.Logger.Error(err)
			return err
		}
		return err
	})
	return regCode, err
}

// DenyRequest : Update status deny request
// Params      : regCodeRepo - RegCodeRepository import , param UpdateRequestStatusParams
// Returns     : register error
func (repo *PgRequestRepository) DenyRequest(updateRequestStatusParams *param.UpdateRequestStatusParams) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		err := repo.UpdateStatusRequestWithTX(tx, updateRequestStatusParams.RequestID, updateRequestStatusParams.Status)
		if err != nil {
			repo.Logger.Error(err)
			return err
		}

		return err
	})
	return err
}

// GetResendRegistCode : Update status deny request
// Params      : regCodeRepo - RegCodeRepository import , param UpdateRequestStatusParams
// Returns     : register error
func (repo *PgRequestRepository) GetResendRegistCode(regCodeRepo rp.RegCodeRepository, resendEmailParams *param.ResendEmailParams) (m.RegistrationCode, error) {
	regCode := m.RegistrationCode{}

	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		regCode, errTx = regCodeRepo.InsertRegCodeWithTx(tx, resendEmailParams.Email, resendEmailParams.RequestID)

		if errTx != nil {
			repo.Logger.Error(errTx)
			return errTx
		}
		return errTx
	})

	return regCode, err
}

// CheckExistRequestUser : check invite email exist on user/request
// Params                : userRepo ,emailAddrs , orgID
// Returns               : error email list
func (repo *PgRequestRepository) CheckExistRequestUser(userRepo rp.UserRepository, emailAddrs []string, orgID int) ([]string, error) {
	var err error
	errorEmails := []string{}
	for _, emailAddr := range emailAddrs {

		requestRegistrationParams := param.RequestRegistrationParams{
			EmailAddr:      emailAddr,
			OrganizationID: orgID,
		}
		registrationRequestObj, err := repo.GetRegRequests(&requestRegistrationParams)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return errorEmails, err
		}

		if registrationRequestObj.ID != 0 {
			errorEmails = append(errorEmails, emailAddr)
			return errorEmails, nil
		}

		user, err := userRepo.GetUserByEmailOrganizationID(emailAddr, orgID)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return errorEmails, err
		}

		if user.ID != 0 {
			errorEmails = append(errorEmails, emailAddr)
			return errorEmails, nil
		}

	}
	return errorEmails, err
}
