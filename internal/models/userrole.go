package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserRole : struct for db table user_roles
type UserRole struct {
	cm.BaseModel

	Name        string
	Description string
}
