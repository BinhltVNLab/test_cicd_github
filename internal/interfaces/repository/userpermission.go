package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	// m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type UserPermissionRepository interface {
	SelectPermissions(organizationId int, userId int) ([]param.SelectPermissionRecords, error)
	UpdateUserPermission(organizationId int, param *param.EditUserPermissionParam) error
	SelectUserPermission(organizationId int, params *param.UserPermissionParams) ([]param.UserPermissionRecords, int, error)
}
