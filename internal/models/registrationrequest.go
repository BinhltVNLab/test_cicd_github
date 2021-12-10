package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// RegistrationRequest : struct for db table (registration_requests)
type RegistrationRequest struct {
	cm.BaseModel

	tableName      struct{}
	ID             int
	Type           int
	Status         int
	Email          string
	OrganizationID int
	Message        string
}
