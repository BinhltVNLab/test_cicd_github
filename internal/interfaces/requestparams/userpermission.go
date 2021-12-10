package requestparams

import (
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"

type CreateUserPermissionParam struct {
	UserID     int `json:"user_id" valid:"required"`
	FunctionID int `json:"function_id" valid:"required"`
	Status     int `json:"status" valid:"required"`
}

type EditUserPermissionParam struct {
	ID         int `json:"user_id" valid:"required"`
	FunctionID int `json:"function_id" valid:"required"`
	Status     int `json:"status" valid:"required"`
}

type SelectUserPermissionParam struct {
	UserId int `json:"user_id" valid:"required"`
}

type RemoveUserPermissionParam struct {
	Id int `json:"id" valid:"required"`
}

type SelectPermissionRecords struct {
	FunctionId int    `json:"function_id"`
	Status     int    `json:"status"`
	Name       string `json:"name"`
	ModuleId   int    `json:"module_id"`
}
type UserPermissionRecords struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	RoleID    int    `json:"roleId"`
	Avatar    string `json:"avatar"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	HasCustom int    `json:"has_custom"`
}
type UserPermissionParams struct {
	Name        string `json:"name" valid:"required"`
	CurrentPage int    `json:"current_page" valid:"required"`
	RowPerPage  int    `json:"row_per_page" valid:"required"`
}

type DataInitOrg struct {
	OrganizationId int           `json:"organization_id" valid:"required"`
	Modules        []m.OrgModule `json:"modules" valid:"required"`
}

type Module struct {
	ModuleId int  `json:"module_id" valid:"required"`
	Status   bool `json:"status" valid:"required"`
}

type FunctionRecord struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type SettingOrgFunctionsParam struct {
	OrganizationId int `json:"organization_id" valid:"required"`
}

type SettingUserFunctionsParam struct {
	OrganizationId int `json:"organization_id" valid:"required"`
	UserID         int `json:"user_id" valid:"required"`
}
