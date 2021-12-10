package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserLeaveBonus : struct for db table technologies
type Technology struct {
	cm.BaseModel

	tableName      struct{}
	Name           string
	OrganizationId int
	Priority       int

	Users []User `pg:"many2many:user_technologies"`
}
