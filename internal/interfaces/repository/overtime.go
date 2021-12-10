package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type OvertimeRepository interface {
	InsertOvertimeRequest(
		createOvertimeParams *param.CreateOvertimeParams,
		organizationId int,
		notificationRepo NotificationRepository,
		userRepo UserRepository,
		uniqueUsersId []int,
	) (string, string, error)
	UpdateStatusOvertimeRequest(
		updateRequestStatusParams *param.UpdateRequestStatusParams,
		notificationRepo NotificationRepository,
		leaveRepo LeaveRepository,
		userRepo UserRepository,
		organizationId int,
		userId int,
		hour float64,
	) (string, string, error)
	SelectOvertimeRequests(
		organizationId int,
		getOvertimeRequestsParams *param.GetOvertimeRequestsParams,
	) ([]param.OvertimeRequestsRecords, int, error)
	SelectOvertimeRequestById(id int) (m.UserOvertimeRequest, error)
	UpdateOvertimeRequest(params *param.UpdateOvertimeRequestParams) error
	InsertOvertimeWeight(
		organizationId int,
		params *param.CreateOvertimeWeightParams,
		settingStep int,
		orgRepo OrgRepository,
	) error
	UpdateOvertimeWeight(params *param.EditOvertimeWeightParams) error
	CountOvertimeWeightByField(field string, value int) (int, error)
	SelectOvertimeWeightByOrganizationId(organizationId int) (m.OvertimeWeight, error)
}
