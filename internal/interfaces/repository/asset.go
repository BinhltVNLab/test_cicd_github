package repository

import (
	"time"

	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type AssetRepository interface {
	SelectAssetList(organizationID int, params *param.GetAssetListParams) ([]param.AssetListRecord, int, error)
	InsertAssetType(organizationID int, params *param.CreateAssetTypeParams) error
	SelectAssetType(organizationID int) ([]m.AssetType, error)
	SelectAssetLog(organizationID int, params *param.GetAssetLogParams) ([]param.AssetLogRecord, int, error)
	InsertAssetRequest(
		organizationID int,
		params *param.CreateRequestAssetParams,
		notificationRepo NotificationRepository,
		userRepo UserRepository,
		uniqueUsersID []int) (string, string, error)
	CreateAsset(organizationID int, params *param.CreateAssetParams) (m.Asset, error)
	SelectRequestAsset(organizationID int, params *param.GetAssetRequestParams) ([]param.RequestAssetRecord, int, error)
	UpdateRequestAsset(organizationID int, params *param.EditAssetRequestParams) error
	UpdateAssetUser(assetId int, userId int, status int, status_req int) error
	DeleteRequestAsset(ID int) error
	CreateAssetLog(organizationID int, assetId int, userId int, startDay time.Time) error
	UpdateAssetLog(organizationID int, currentUserID int, assetID int, endDayUsing time.Time) (int, error)
	UpdateAsset(organizationID int, params *param.EditAssetParams) error
	SelectAssetByID(ID int) (m.Asset, error)
	DeleteAsset(ID int) error
	DeleteAssetType(ID int) error
	SelectAssetByOrgID(organizationID int) ([]m.Asset, error)
	ImportAsset(organizationID int, assetArray []m.Asset) error
	InsertAssetLogWithTx(tx *pg.Tx, organizationID int,
		assetId int,
		userId int,
		startDay time.Time) (m.AssetLog, error)
}
