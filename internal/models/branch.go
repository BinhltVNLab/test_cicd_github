package models

import cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"

type Branch struct {
	cm.BaseModel

	tableName      struct{}
	Name           string
	OrganizationId int
	Priority       int
}
