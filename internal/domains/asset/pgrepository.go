package asset

import (
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type PgAssetRepository struct {
	cm.AppRepository
}

func NewPgAssetRepository(logger echo.Logger) (repo *PgAssetRepository) {
	repo = &PgAssetRepository{}
	repo.Init(logger)
	return
}

func (repo *PgAssetRepository) SelectAssetList(organizationId int, getAssetListParams *param.GetAssetListParams) ([]param.AssetListRecord, int, error) {
	var records []param.AssetListRecord
	q := repo.DB.Model(&m.Asset{})

	q.Column("asset.id", "asset.asset_name", "asset.asset_code", "asset.branch_id", "asset.user_id", "asset.status",
		"asset.description", "asset.date_started_use", "asset.license_end_date",
		"asset.purchase_price", "asset.managed_by", "asset.date_of_purchase", "asset.depreciation_period").
		ColumnExpr("ast.name as asset_type").ColumnExpr("ast.id as asset_type_id").
		ColumnExpr("asset.status_req as status_req").
		ColumnExpr("asset.created_at").
		Join("JOIN asset_types as ast on asset.asset_type_id = ast.id").
		Join("FULL OUTER JOIN user_profiles as up on up.user_id = asset.user_id").
		Where("asset.organization_id = ?", organizationId)

	if getAssetListParams.AssetName != "" {
		q.Where("LOWER(asset.asset_name) LIKE LOWER(?)", "%"+getAssetListParams.AssetName+"%")
	}

	if getAssetListParams.AssetCode != "" {
		q.Where("LOWER(asset.asset_code) LIKE LOWER(?)", "%"+getAssetListParams.AssetCode+"%")
	}

	if getAssetListParams.AssetTypeID != 0 {
		q.Where("asset.asset_type_id = ?", getAssetListParams.AssetTypeID)
	}

	if getAssetListParams.BranchID != 0 {
		q.Where("asset.branch_id = ?", getAssetListParams.BranchID)
	}

	if getAssetListParams.UserName != "" {
		userName := "%" + getAssetListParams.UserName + "%"
		q.Where("vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if getAssetListParams.Status != 0 {
		q.Where("asset.status = ?", getAssetListParams.Status)
	}

	q.Offset((getAssetListParams.CurrentPage - 1) * getAssetListParams.RowPerPage).
		Order("asset.created_at DESC").
		
		Limit(getAssetListParams.RowPerPage)

	totalRow, err := q.SelectAndCount(&records)

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return records, totalRow, err
}

func (repo *PgAssetRepository) InsertAssetType(
	organizationId int,
	params *param.CreateAssetTypeParams,
) error {
	assetType := m.AssetType{
		Name:           params.Name,
		OrganizationID: organizationId,
	}

	err := repo.DB.Insert(&assetType)

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) SelectAssetLog(OrgID int, getAssetLogParams *param.GetAssetLogParams) ([]param.AssetLogRecord, int, error) {
	var records []param.AssetLogRecord
	q := repo.DB.Model(&m.AssetLog{})

	q.Column("asset_logs.asset_id", "asset_logs.user_id", "asset_logs.start_day_using", "asset_logs.end_day_using", "assets.asset_name", "assets.date_started_use",
		"assets.asset_code", "assets.status", "assets.branch_id", "up.first_name", "up.last_name").
		ColumnExpr("ast.name as asset_type_name").
		Join("JOIN assets on assets.id = asset_logs.asset_id").
		Join("JOIN asset_types as ast on assets.asset_type_id = ast.id").
		Join("FULL OUTER JOIN user_profiles as up on up.user_id = asset_logs.user_id").
		Where("asset_logs.organization_id = ?", OrgID)

	if getAssetLogParams.AssetName != "" {
		q.Where("LOWER(assets.asset_name) LIKE LOWER(?)", "%"+getAssetLogParams.AssetName+"%")
	}

	if getAssetLogParams.AssetCode != "" {
		q.Where("LOWER(assets.asset_code) LIKE LOWER(?)", "%"+getAssetLogParams.AssetCode+"%")
	}

	if getAssetLogParams.AssetTypeId != 0 {
		q.Where("assets.asset_type_id = ?", getAssetLogParams.AssetTypeId)
	}

	if getAssetLogParams.BranchID != 0 {
		q.Where("assets.branch_id = ?", getAssetLogParams.BranchID)
	}

	if getAssetLogParams.UserName != "" {
		userName := "%" + getAssetLogParams.UserName + "%"
		q.Where("vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if getAssetLogParams.Status != 0 {
		q.Where("assets.status = ?", getAssetLogParams.Status)
	}

	q.Offset((getAssetLogParams.CurrentPage - 1) * getAssetLogParams.RowPerPage).
		Order("asset_logs.created_at DESC").
		Limit(getAssetLogParams.RowPerPage)

	totalRow, err := q.SelectAndCount(&records)

	return records, totalRow, err
}

func (repo *PgAssetRepository) SelectAssetType(organizationId int) ([]m.AssetType, error) {
	var AssetTypeList []m.AssetType
	err := repo.DB.Model(&AssetTypeList).
		Column("id", "name").
		Where("organization_id = ?", organizationId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return AssetTypeList, err
}

func (repo *PgAssetRepository) InsertAssetRequest(
	organizationId int,
	createRequestAssetParams *param.CreateRequestAssetParams,
	notificationRepo rp.NotificationRepository,
	userRepo rp.UserRepository,
	uniqueUsersID []int,
) (string, string, error) { 
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		userAssetRequest := m.UserAssetRequest{
			OrganizationID: organizationId,
			CreatedBy:      createRequestAssetParams.UserID,
			AssetID:        createRequestAssetParams.AssetId,
			Status:         createRequestAssetParams.StatusReq,
		}

		transErr = tx.Insert(&userAssetRequest)
		if transErr != nil {
			return transErr
		}

		notificationParams := new(param.InsertNotificationParam)
		notificationParams.Content = "has just created a asset request"
		notificationParams.RedirectUrl = "/hrm/asset/manage-asset-request?id=" + strconv.Itoa(userAssetRequest.ID)

		for _, userID := range uniqueUsersID {
			if userID == createRequestAssetParams.UserID {
				continue
			}
			notificationParams.Receiver = userID
			transErr = notificationRepo.InsertNotificationWithTx(tx, organizationId, createRequestAssetParams.UserID, notificationParams)
			if transErr != nil {
				return transErr
			}
		}

		body = notificationParams.Content
		link = notificationParams.RedirectUrl

		return transErr
	})

	return body, link, err
}

func (repo *PgAssetRepository) CreateAsset(orgID int, createAssetParams *param.CreateAssetParams) (m.Asset, error) {
	var dateOfPurchase time.Time
	if createAssetParams.DateOfPurchase != "" {
		dateOfPurchase = calendar.ParseTime(cf.FormatDateDatabase, createAssetParams.DateOfPurchase)
	}
	var licenseEndDate time.Time
	if createAssetParams.LicenseEndDate != "" {
		licenseEndDate = calendar.ParseTime(cf.FormatDateDatabase, createAssetParams.LicenseEndDate)
	}
	var dateStartedUse time.Time
	if createAssetParams.DateStartedUse != "" {
		dateStartedUse = calendar.ParseTime(cf.FormatDateDatabase, createAssetParams.DateStartedUse)
	}

	asset := m.Asset{
		OrganizationId:     orgID,
		UserId:             createAssetParams.UserId,
		AssetTypeId:        createAssetParams.AssetTypeId,
		BranchId:           createAssetParams.BrandId,
		ManagedBy:          createAssetParams.ManagedBy,
		AssetName:          createAssetParams.AssetName,
		AssetCode:          createAssetParams.AssetCode,
		Description:        createAssetParams.Description,
		Status:             createAssetParams.Status,
		PurchasePrice:      createAssetParams.PurchasePrice,
		DepreciationPeriod: createAssetParams.DepreciationPeriod,
		DateOfPurchase:     dateOfPurchase,
		LicenseEndDate:     licenseEndDate,
		DateStartedUse:     dateStartedUse,
	}

	err := repo.DB.Insert(&asset)

	if err != nil {
		repo.Logger.Error(err)
	}
	return asset, err
}

func (repo *PgAssetRepository) SelectRequestAsset(OrganizationID int, getAssetRequestParams *param.GetAssetRequestParams) ([]param.RequestAssetRecord, int, error) {
	var records []param.RequestAssetRecord
	q := repo.DB.Model(&m.UserAssetRequest{})

	q.Column("asset.asset_code", "asset.asset_name", "asset.description", "uar.status", "uar.id").
		ColumnExpr("b.name as branch").
		ColumnExpr("at.name as asset_type").
		ColumnExpr("uar.created_by created_by").
		ColumnExpr("asset.id as asset_id").
		ColumnExpr("asset.managed_by as managed_by").
		Join("JOIN assets AS asset on asset.id = uar.asset_id").
		Join("JOIN asset_types AS at on asset.asset_type_id = at.id").
		Join("JOIN branches AS b on b.id = asset.branch_id").
		Where("uar.organization_id = ?", OrganizationID)

	if getAssetRequestParams.AssetType != 0 {
		q.Where("at.id = ?", getAssetRequestParams.AssetType)
	}

	if getAssetRequestParams.BranchID != 0 {
		q.Where("up.branch = ?", getAssetRequestParams.BranchID)
	}

	if getAssetRequestParams.UserName != "" {
		userName := "%" + getAssetRequestParams.UserName + "%"
		q.Where("vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	q.Offset((getAssetRequestParams.CurrentPage - 1) * getAssetRequestParams.RowPerPage).
		Order("uar.created_at DESC").
		Limit(getAssetRequestParams.RowPerPage)

	totalRow, err := q.SelectAndCount(&records)

	return records, totalRow, err
}

func (repo *PgAssetRepository) UpdateRequestAsset(organizationId int, params *param.EditAssetRequestParams) error {
	records := m.UserAssetRequest{
		Status: params.Status,
	}

	_, err := repo.DB.Model(&records).
		Where("id = ?", params.ID).
		Where("organization_id = ?", organizationId).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) UpdateAssetUser(assetId int, userId int, status int, status_req int) error {
	records := m.Asset{
		UserId: userId,
		Status: status,
		StatusReq: status_req,
	}

	_, err := repo.DB.Model(&records).
	    Column("user_id", "status", "status_req").
		Where("id = ?", assetId).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) UpdateAsset(orgID int, editAssetParams *param.EditAssetParams) error {
	var dateOfPurchase time.Time
	if editAssetParams.DateOfPurchase != "" {
		dateOfPurchase = calendar.ParseTime(cf.FormatDateDatabase, editAssetParams.DateOfPurchase)
	}
	var licenseEndDate time.Time
	if editAssetParams.LicenseEndDate != "" {
		licenseEndDate = calendar.ParseTime(cf.FormatDateDatabase, editAssetParams.LicenseEndDate)
	}
	var dateStartedUse time.Time
	if editAssetParams.DateStartedUse != "" {
		dateStartedUse = calendar.ParseTime(cf.FormatDateDatabase, editAssetParams.DateStartedUse)
	}

	asset := m.Asset{
		UserId:             editAssetParams.UserId,
		AssetTypeId:        editAssetParams.AssetTypeId,
		BranchId:           editAssetParams.BrandId,
		ManagedBy:          editAssetParams.ManagedBy,
		AssetName:          editAssetParams.AssetName,
		AssetCode:          editAssetParams.AssetCode,
		Description:        editAssetParams.Description,
		Status:             editAssetParams.Status,
		PurchasePrice:      editAssetParams.PurchasePrice,
		DepreciationPeriod: editAssetParams.DepreciationPeriod,
		DateOfPurchase:     dateOfPurchase,
		LicenseEndDate:     licenseEndDate,
		DateStartedUse:     dateStartedUse,
	}

	_, err := repo.DB.Model(&asset).
	    Column("user_id", "status", "asset_type_id", "branch_id", "managed_by", "asset_name", "asset_code", "description", "purchase_price",
		"depreciation_period", "date_of_purchase", "license_end_date", "date_started_use").
		Where("id = ?", editAssetParams.ID).
		Where("organization_id = ?", orgID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) UpdateAssetLog(organizationID int, currentUserID int, assetID int, endDayUsing time.Time) (int, error) {
	records := m.AssetLog{
		EndDayUsing: endDayUsing,
	}

	result, err := repo.DB.Model(&records).
		Where("user_id = ? and asset_id = ?", currentUserID, assetID).
		Where("organization_id = ?", organizationID).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
	}

	return result.RowsAffected(), err
}

func (repo *PgAssetRepository) DeleteRequestAsset(ID int) error {
	_, err := repo.DB.Model(&m.UserAssetRequest{}).
		Where("id = ?", ID).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) CreateAssetLog(
	organizationID int,
	assetId int,
	userId int,
	startDay time.Time,
) error {
	assetLog := m.AssetLog{
		OrganizationID: organizationID,
		AssetId:        assetId,
		UserId:         userId,
		StartDayUsing:  startDay,
	}

	err := repo.DB.Insert(&assetLog)
	return err
}

func (repo *PgAssetRepository) InsertAssetLogWithTx(tx *pg.Tx, organizationID int,
	assetId int,
	userId int,
	startDay time.Time) (m.AssetLog, error) {
	assetLog := m.AssetLog{
		OrganizationID: organizationID,
		AssetId:        assetId,
		UserId:         userId,
		StartDayUsing:  startDay,
	}

	err := tx.Insert(&assetLog)

	if err != nil {
		repo.Logger.Error(err)
	}

	return assetLog, err
}

func (repo *PgAssetRepository) SelectAssetByID(Id int) (m.Asset, error) {
	var asset m.Asset
	err := repo.DB.Model(&asset).
		Column("asset.asset_type_id", "asset.asset_code", "asset.asset_name", "asset.managed_by",
			"asset.user_id", "asset.branch_id", "asset.status", "asset.description", "asset.purchase_price",
			"asset.date_of_purchase", "asset.license_end_date", "asset.date_started_use", "asset.depreciation_period").
		Where("asset.id = ?", Id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return asset, err
}

// DeleteAsset : Remove asset
func (repo *PgAssetRepository) DeleteAsset(ID int) error {
	_, err := repo.DB.Model(&m.Asset{}).
		Where("id = ?", ID).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// DeleteAssetType : Remove asset
func (repo *PgAssetRepository) DeleteAssetType(ID int) error {
	_, err := repo.DB.Model(&m.AssetType{}).
		Where("id = ?", ID).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgAssetRepository) SelectAssetByOrgID(organizationID int) ([]m.Asset, error) {
	var assetList []m.Asset
	err := repo.DB.Model(&assetList).
		Where("asset.organization_id = ?", organizationID).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return assetList, err
}

func (repo *PgAssetRepository) ImportAsset(organizationID int, assetArray []m.Asset) error {
	assetList, err := repo.SelectAssetByOrgID(organizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return err
	}

	err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error = nil

		for _, itemImportAsset := range assetArray {
			changeUserAsset := false
			for _, assetDBItem := range assetList {
				if itemImportAsset.AssetCode == assetDBItem.AssetCode {
					itemImportAsset.ID = assetDBItem.ID

					if itemImportAsset.UserId > 0 && itemImportAsset.UserId != assetDBItem.UserId {
						changeUserAsset = true
					}
					break
				}
			}

			if itemImportAsset.ID > 0 {
				_, errTx = tx.Model(&itemImportAsset).
					Column(
						"updated_at",
						"user_id",
						"asset_type_id",
						"asset_name",
						"description",
						"branch_id",
						"managed_by",
						"date_of_purchase",
						"status",
						"depreciation_period",
						"purchase_price",
						"license_end_date",
						"date_started_use",
					).
					Where("id = ?", itemImportAsset.ID).
					Update()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				if itemImportAsset.UserId > 0 {
					assetLog := m.AssetLog{
						OrganizationID: organizationID,
						AssetId:        itemImportAsset.ID,
						UserId:         itemImportAsset.UserId,
						StartDayUsing:  utils.TimeNowUTC(),
					}

					_, errTx := tx.Model(&assetLog).Insert()

					if errTx != nil {
						repo.Logger.Error(errTx)
						return errTx
					}
				}

				if changeUserAsset && itemImportAsset.UserId > 0 {
					repo.InsertAssetLogWithTx(tx, organizationID, itemImportAsset.ID, itemImportAsset.UserId, utils.TimeNowUTC())
				}
			} else {
				_, errTx = tx.Model(&itemImportAsset).Insert()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				if itemImportAsset.UserId > 0 {
					repo.InsertAssetLogWithTx(tx, organizationID, itemImportAsset.ID, itemImportAsset.UserId, utils.TimeNowUTC())
				}
			}
		}

		return errTx
	})

	if err != nil {
		repo.Logger.Error(err)
		return err
	}

	return err
}
