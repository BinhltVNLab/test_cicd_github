package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type AdminRepository interface {
	InsertOrgModule(params *param.DataInitOrg) error
	SelectFunctionsOrg(organizationId int) ([]param.FunctionRecord, error)
	InsertUserPermission(organizationId int, userID int, functionId int, status int) error
	SelectModules() ([]m.Module , error)
}
