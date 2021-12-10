package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserProject : struct for db table user_projects
type UserProject struct {
	cm.BaseModel

	UserID    int
	ProjectID int
}
