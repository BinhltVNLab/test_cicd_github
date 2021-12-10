package models

import cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"

type OrganizationModule struct {
	cm.BaseModel

	TableName      struct{} `sql:"alias:orgm"`
	OrganizationId int
	Modules        []OrgModule
}

type OrgModule struct {
	MoudleID int  `json:"module_id" valid:"required"`
	Status   bool `json:"status" valid:"required"`
}
