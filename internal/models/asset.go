package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

type Asset struct {
	cm.BaseModel

	TableName          struct{} `sql:"alias:asset"`
	OrganizationId     int
	UserId             int
	AssetTypeId        int
	BranchId           int
	ManagedBy          int
	AssetName          string
	AssetCode          string
	Description        string
	Status             int
	PurchasePrice      float64
	DateOfPurchase     time.Time
	DepreciationPeriod int
	LicenseEndDate     time.Time
	DateStartedUse     time.Time
	StatusReq          int
}

type AssetType struct {
	cm.BaseModel

	TableName      struct{}
	OrganizationID int
	Name           string
}

type UserAssetRequest struct {
	cm.BaseModel

	TableName      struct{} `sql:"alias:uar"`
	OrganizationID int
	AssetID        int
	CreatedBy      int
	Status         int
}
