package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// Organization : struct for db table organizations
type Organization struct {
	cm.BaseModel

	tableName             struct{} `sql:"alias:org"`
	Name                  string
	Tag                   string
	PhoneNumber           string
	Address               string
	Description           string
	Email                 string
	EmailPassword         string
	SettingStep           int
	ExpirationResetDayOff int
}
