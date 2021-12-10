package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// Project : struct for db table (projects)
type Project struct {
	cm.BaseModel
	tableName      struct{} `sql:"alias:prj"`
	OrganizationID int
	Name           string
	Description    string
	Targets        []Target
	ManagedBy      int
}

// Target jsonb
type Target struct {
	Year          int    `json:"year" valid:"required"`
	Quarter       int    `json:"quarter" valid:"required,range(1|4)"`
	TargetContent string `json:"content" valid:"required"`
}
