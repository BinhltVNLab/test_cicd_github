package models

import cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"

type Module struct {
	cm.BaseModel

	TableName struct{} `sql:"alias:module"`
	ID        int
	Name      string
}
