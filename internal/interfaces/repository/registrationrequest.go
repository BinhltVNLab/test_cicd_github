package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// RegistRequestRepository interface
type RegistRequestRepository interface {
	GetListRequest(requestRegistrationParams *param.ManageRequestParams) ([]m.RegistrationRequest, int, error)
	GetRegRequests(requestRegistrationParams *param.RequestRegistrationParams) (m.RegistrationRequest, error)
	GetRegRequestsByID(id int) (m.RegistrationRequest, error)
	InsertRegRequest(requestType int, requestStatus int, requestRegistrationParams *param.RequestRegistrationParams) (m.RegistrationRequest, error)
	InsertRegRequestWithTx(tx *pg.Tx, requestType int, requestStatus int, requestEmail string, orgID int, message string) (m.RegistrationRequest, error)
	SaveAcceptedRegRequest(
		regCodeRepo RegCodeRepository,
		requestEmail []string,
		organizationID int,
		message string,
	) ([]int, []int, map[string]string, error)
	UpdateStatusRequestWithTX(tx *pg.Tx, rquestID int, status int) error
	AcceptRequest(regCodeRepo RegCodeRepository, updateRequestStatusParams *param.UpdateRequestStatusParams) (m.RegistrationCode, error)
	DenyRequest(updateRequestStatusParams *param.UpdateRequestStatusParams) error
	GetResendRegistCode(regCodeRepo RegCodeRepository, resendEmailParams *param.ResendEmailParams) (m.RegistrationCode, error)
	CheckExistRequestUser(userRepo UserRepository, emailAddrs []string, orgID int) ([]string, error)
	DeleteInviteRequests(regRequestIds []int, regCodeIds []int) error
}
