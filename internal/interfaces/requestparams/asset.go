package requestparams

import "time"

// GetAssetListParams struct for receive param from frontend
type GetAssetListParams struct {
	AssetName   string `json:"asset_name"`
	AssetCode   string `json:"asset_code"`
	AssetTypeID int    `json:"asset_type_id"`
	BranchID    int    `json:"branch_id"`
	UserName    string `json:"user_name"`
	Status      int    `json:"status"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// AssetListRecord struct for receive param from frontend
type AssetListRecord struct {
	ID                 int       `json:"id"`
	AssetName          string    `json:"asset_name"`
	AssetCode          string    `json:"asset_code"`
	AssetType          string    `json:"asset_type"`
	AssetTypeId        int       `json:"asset_type_id"`
	BranchID           int       `json:"branch_id"`
	UserID             int       `json:"user_id"`
	Status             int       `json:"status"`
	StatusReq          int       `json:"status_req"`
	Description        string    `json:"description"`
	DateStartedUse     time.Time `json:"date_started_use"`
	LicenseEndDate     time.Time `json:"license_end_date"`
	DateOfPurchase     time.Time `json:"date_of_purchase"`
	PurchasePrice      int       `json:"purchase_price"`
	ManagedBy          int       `json:"managed_by"`
	CreatedAt	   	time.Time 	`json:"created_at"`
	DepreciationPeriod int       `json:"depreciation_period"`
}

// CreateAssetTypeParams struct for receive param from frontend
type CreateAssetTypeParams struct {
	Name string `json:"name" valid:"required"`
}

// GetAssetLogParams struct for receive param from frontend
type GetAssetLogParams struct {
	AssetName   string `json:"asset_name"`
	AssetCode   string `json:"asset_code"`
	AssetTypeId int    `json:"asset_type_id"`
	BranchID    int    `json:"branch_id"`
	UserName    string `json:"user_name"`
	Status      int    `json:"status"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// AssetLogRecord struct for receive param from frontend
type AssetLogRecord struct {
	AssetId       int       `json:"asset_id"`
	AssetName     string    `json:"asset_name"`
	AssetCode     string    `json:"asset_code"`
	AssetTypeName string    `json:"asset_type_name"`
	BranchID      int       `json:"branch_id"`
	UserID        int       `json:"user_id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Status        int       `json:"status"`
	DateStartedUse     time.Time `json:"date_started_use"`
	StartDayUsing time.Time `json:"start_day_using"`
	EndDayUsing   time.Time `json:"end_day_using"`
}

// CreateRequestAccessParams struct for receive param from frontend
type CreateRequestAssetParams struct {
	AssetId     int `json:"asset_id"`
	UserID      int `json:"user_id"`
	StatusAsset int `json:"status"`
	StatusReq   int `json:"status_req"`
}

// CreateAssetParams struct for receive param from frontend
type CreateAssetParams struct {
	AssetTypeId        int     `json:"asset_type_id" valid:"required"`
	UserId             int     `json:"user_id"`
	BrandId            int     `json:"branch_id" valid:"required"`
	AssetCode          string  `json:"asset_code" valid:"required"`
	AssetName          string  `json:"asset_name" valid:"required"`
	ManagedBy          int     `json:"managed_by" valid:"required"`
	Status             int     `json:"status" valid:"required"`
	Description        string  `json:"description"`
	PurchasePrice      float64 `json:"purchase_price"`
	DepreciationPeriod int     `json:"depreciation_period"`
	DateOfPurchase     string  `json:"date_of_purchase"`
	LicenseEndDate     string  `json:"license_end_date"`
	DateStartedUse     string  `json:"date_started_use"`
	StatusReq          int     `json:"status_req"`
}

// RequestAssetRecord struct for receive param from frontend
type RequestAssetRecord struct {
	ID          int    `json:"id"`
	AssetID     int    `json:"asset_id"`
	AssetName   string `json:"asset_name"`
	AssetCode   string `json:"asset_code"`
	AssetType   string `json:"asset_type"`
	Description string `json:"description"`
	CreatedBy   int    `json:"created_by"`
	Branch      string `json:"branch"`
	Status      int    `json:"status"`
	ManagedBy   int    `json:"managed_by"`
}

// GetAssetRequestParams struct for receive param from frontend
type GetAssetRequestParams struct {
	AssetType   int    `json:"asset_type"`
	BranchID    int    `json:"branch_id"`
	UserName    string `json:"user_name"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// EditAssetRequestParams struct for receive param from frontend
type EditAssetRequestParams struct {
	ID      int `json:"id"`
	UserID  int `json:"user_id"`
	AssetID int `json:"asset_id"`
	Status  int `json:"status"`
}

type EditAssetParams struct {
	ID                 int     `json:"id" valid:"required"`
	AssetTypeId        int     `json:"asset_type_id" valid:"required"`
	UserId             int     `json:"user_id"`
	BrandId            int     `json:"branch_id" valid:"required"`
	AssetCode          string  `json:"asset_code" valid:"required"`
	AssetName          string  `json:"asset_name" valid:"required"`
	ManagedBy          int     `json:"managed_by" valid:"required"`
	Status             int     `json:"status" valid:"required"`
	Description        string  `json:"description"`
	PurchasePrice      float64 `json:"purchase_price"`
	DepreciationPeriod int     `json:"depreciation_period"`
	DateOfPurchase     string  `json:"date_of_purchase"`
	LicenseEndDate     string  `json:"license_end_date"`
	DateStartedUse     string  `json:"date_started_use"`
}

// CreateRequestAccessParams struct for receive param from frontend
type GetAssetParams struct {
	ID int `json:"id" valid:"required"`
}

type DeleteParams struct {
	ID int `json:"id" valid:"required"`
}
type CreateAssetLogParams struct {
	AssetID       int `json:"asset_id" valid:"required"`
	CurrentUserID int `json:"current_user_id"`
	NewUserID     int `json:"new_user_id"`
}
