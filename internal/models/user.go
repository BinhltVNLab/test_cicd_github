package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// User : struct for db table users
type User struct {
	cm.BaseModel

	tableName                struct{} `sql:"alias:usr"`
	OrganizationID           int
	Email                    string
	Password                 string
	RoleID                   int
	LastLoginTime            time.Time
	ResetPasswordCode        string
	CodeExpiredAt            time.Time
	EmailForUpdate           string
	UpdateEmailCode          string
	UpdateEmailCodeExpiredAt time.Time
	GoogleID                 string
	LanguageId               int

	UserProfile      UserProfile
	Role             UserRole `pg:",fk:role_id"`
	Organization     Organization
	UserTimekeeping  []UserTimekeeping  `pg:",fk:user_id"`
	TargetEvaluation []TargetEvaluation `pg:",fk:user_id"`

	UserTechnologies []UserTechnology `pg:"many2many:user_technologies"`
}
