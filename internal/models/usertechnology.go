package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// UserTechnology : struct for db table user_technologies
type UserTechnology struct {
	cm.BaseModel

	UserID       int
	TechnologyId int

	User       *User
	Technology *Technology
}
