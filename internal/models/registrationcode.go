package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// RegistrationCode : struct for db table (registration_codes)
type RegistrationCode struct {
	cm.BaseModel

	tableName             struct{}
	Email                 string
	Code                  string
	RegistrationRequestID int
	ExpiredAt             time.Time
	GoogleID              string
}
