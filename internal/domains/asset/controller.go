package asset

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	ex "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/excel"
)

type Controller struct {
	cm.BaseController
	afb.FirebaseCloudMessage

	assetRepository  rp.AssetRepository
	UserRepo         rp.UserRepository
	BranchRepo       rp.BranchRepository
	NotificationRepo rp.NotificationRepository
	FcmTokenRepo     rp.FcmTokenRepository
}

func NewAssetController(logger echo.Logger, assetRepository rp.AssetRepository, userRepo rp.UserRepository,
	branchRepo rp.BranchRepository, notificationRepo rp.NotificationRepository, fcmTokenRepo rp.FcmTokenRepository) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, afb.FirebaseCloudMessage{}, assetRepository,
		userRepo, branchRepo, notificationRepo, fcmTokenRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) GetAssetList(c echo.Context) error {
	getAssetListParams := new(param.GetAssetListParams)

	if err := c.Bind(getAssetListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getAssetListParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	assetRecords, totalRow, err := ctr.assetRepository.SelectAssetList(
		userProfile.OrganizationID,
		getAssetListParams,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var assetList []map[string]interface{}
	var dateStartedUse string
	var depreciationEndDate string
	var depreciationRates float64 = 0
	var licenseEndDate string

	for _, record := range assetRecords {
		if record.DateStartedUse.Format(cf.FormatDateDisplay) == "0001/01/01" {
			dateStartedUse = ""
			depreciationEndDate = ""
		} else {
			dateStartedUse = record.DateStartedUse.Format(cf.FormatDateDisplay)
			if (record.DepreciationPeriod == 0) {
				depreciationRates = 0
			} else {
				depreciationEndDate = record.DateStartedUse.AddDate(0, record.DepreciationPeriod*12, 0).Format(cf.FormatDateDisplay)
				usedMonths := monthsCountSince(record.DateStartedUse)
				depreciationRates = 100 * float64(usedMonths) / float64(record.DepreciationPeriod*12)
			}
		}

		if record.LicenseEndDate.Format(cf.FormatDateDisplay) == "0001/01/01" {
			licenseEndDate = ""
		} else {
			licenseEndDate = record.LicenseEndDate.Format(cf.FormatDateDisplay)
		}
		res := map[string]interface{}{
			"asset_id":              record.ID,
			"asset_name":            record.AssetName,
			"asset_code":            record.AssetCode,
			"asset_type":            record.AssetType,
			"asset_type_id":         record.AssetTypeId,
			"branch_id":             record.BranchID,
			"user_id":               record.UserID,
			"status":                record.Status,
			"status_req":            record.StatusReq,
			"description":           record.Description,
			"date_started_use":      dateStartedUse,
			"created_at":         	 record.CreatedAt.Format(cf.FormatDateDisplay),
			"license_end_date":      licenseEndDate,
			"date_of_purchase":      record.DateOfPurchase.Format(cf.FormatDateDisplay),
			"purchase_price":        record.PurchasePrice,
			"managed_by":            record.ManagedBy,
			"depreciation_period":   record.DepreciationPeriod,
			"depreciation_end_date": depreciationEndDate,
			"depreciation_rates":    toFixed(depreciationRates, 2),
		}
		assetList = append(assetList, res)
	}

	pagination := map[string]interface{}{
		"current_page": getAssetListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getAssetListParams.RowPerPage,
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			users[user.UserID] = user.FullName
		}
	}

	branches, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branchList := make(map[int]string)
	for _, record := range branches {
		branchList[record.Id] = record.Name
	}

	responseData := map[string]interface{}{
		"pagination":           pagination,
		"asset_list":           assetList,
		"users":                users,
		"branches":             branchList,
		"asset_status":         cf.AssetStatus,
		"asset_request_status": cf.RequestAssetStatus,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get asset list successfully.",
		Data:    responseData,
	})
}

func (ctr *Controller) CreateAssetType(c echo.Context) error {
	createAssetTypeParams := new(param.CreateAssetTypeParams)

	if err := c.Bind(createAssetTypeParams); err != nil || strings.TrimSpace(createAssetTypeParams.Name) == "" {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createAssetTypeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err = ctr.assetRepository.InsertAssetType(
		userProfile.OrganizationID,
		createAssetTypeParams,
	)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates") {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Duplicate asset code value violates",
			Data:    err,
		})
	}

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create asset type successfully.",
	})
}

func (ctr *Controller) GetAssetLog(c echo.Context) error {
	getAssetLogParams := new(param.GetAssetLogParams)

	if err := c.Bind(getAssetLogParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getAssetLogParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	assetLogRecords, totalRow, err := ctr.assetRepository.SelectAssetLog(userProfile.OrganizationID, getAssetLogParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			users[user.UserID] = user.FullName
		}
	}

	branches, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branchList := make(map[int]string)
	for _, record := range branches {
		branchList[record.Id] = record.Name
	}

	pagination := map[string]interface{}{
		"current_page": getAssetLogParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getAssetLogParams.RowPerPage,
	}

	responseData := map[string]interface{}{
		"pagination":      pagination,
		"asset_histories": assetLogRecords,
		"users":           users,
		"branches":        branchList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get asset history successfully.",
		Data:    responseData,
	})
}

func (ctr *Controller) GetAssetTypeList(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	assetTypeList, err := ctr.assetRepository.SelectAssetType(userProfile.OrganizationID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}
	var responses []map[string]interface{}

	for _, record := range assetTypeList {
		res := map[string]interface{}{
			"id":   record.ID,
			"name": record.Name,
		}
		responses = append(responses, res)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get asset type list successfully.",
		Data:    responses,
	})
}
func (ctr *Controller) CreateAsset(c echo.Context) error {
	createAssetParams := new(param.CreateAssetParams)

	if err := c.Bind(createAssetParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	_, err := valid.ValidateStruct(createAssetParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	asset, err := ctr.assetRepository.CreateAsset(userProfile.OrganizationID, createAssetParams)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates") {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Duplicate asset code value violates",
			Data:    err,
		})
	}

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if createAssetParams.UserId != 0 {
		now := time.Now()
		err = ctr.assetRepository.CreateAssetLog(userProfile.OrganizationID, asset.ID,
			asset.UserId, now)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create asset successfully.",
		Data:    asset,
	})
}

func (ctr *Controller) GetUserRequestAsset(c echo.Context) error {
	getAssetRequestParams := new(param.GetAssetRequestParams)

	if err := c.Bind(getAssetRequestParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getAssetRequestParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	assetRequestRecords, totalRow, err := ctr.assetRepository.SelectRequestAsset(userProfile.OrganizationID, getAssetRequestParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	pagination := map[string]interface{}{
		"current_page": getAssetRequestParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getAssetRequestParams.RowPerPage,
	}

	userRecords, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			users[user.UserID] = user.FullName
		}
	}

	branches, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branchList := make(map[int]string)
	for _, record := range branches {
		branchList[record.Id] = record.Name
	}

	responseData := map[string]interface{}{
		"pagination":     pagination,
		"asset_requests": assetRequestRecords,
		"users":          users,
		"branches":       branchList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get asset requests successfully.",
		Data:    responseData,
	})
}

func (ctr *Controller) EditUserRequestAsset(c echo.Context) error {
	editAssetRequestParams := new(param.EditAssetRequestParams)

	if err := c.Bind(editAssetRequestParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	_, err := valid.ValidateStruct(editAssetRequestParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.assetRepository.UpdateRequestAsset(userProfile.OrganizationID, editAssetRequestParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	now := time.Now()
	if editAssetRequestParams.Status == cf.AcceptRequestBorrowAsset || editAssetRequestParams.Status == cf.DenyRequestReturn {
		err = ctr.assetRepository.UpdateAssetUser(editAssetRequestParams.AssetID, editAssetRequestParams.UserID, 2, cf.AcceptRequestBorrowAsset)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}

		if editAssetRequestParams.Status == cf.AcceptRequestBorrowAsset {
			err = ctr.assetRepository.CreateAssetLog(userProfile.OrganizationID, editAssetRequestParams.AssetID,
				editAssetRequestParams.UserID, now)
	
			if err != nil && err.Error() != pg.ErrNoRows.Error() {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System Error",
					Data:    err,
				})
			}
		}
	}

	if (editAssetRequestParams.Status == cf.AcceptRequestReturnAsset || editAssetRequestParams.Status == cf.DenyRequestBorrow) {
		err = ctr.assetRepository.UpdateAssetUser(editAssetRequestParams.AssetID, 0, 1, 0)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}

		if editAssetRequestParams.Status == cf.AcceptRequestReturnAsset {
			_, err := ctr.assetRepository.UpdateAssetLog(userProfile.OrganizationID, editAssetRequestParams.UserID, editAssetRequestParams.AssetID, now)
			if err != nil && err.Error() != pg.ErrNoRows.Error() {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System Error",
					Data:    err,
				})
			}
		}
		
	}

	if (editAssetRequestParams.Status == cf.DenyRequestBorrow || editAssetRequestParams.Status == cf.DenyRequestReturn) {
		err = ctr.assetRepository.DeleteRequestAsset(editAssetRequestParams.ID)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update status request successfully.",
	})
}

func (ctr *Controller) CreateRequestAsset(c echo.Context) error {
	createRequestAssetParams := new(param.CreateRequestAssetParams)

	if err := c.Bind(createRequestAssetParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createRequestAssetParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	usersIdGm, err := ctr.UserRepo.SelectIdsOfGM(userProfile.OrganizationID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if createRequestAssetParams.UserID != userProfile.UserProfile.UserID {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "You not have permission to create asset request",
		})
	}

	_, _, err = ctr.assetRepository.InsertAssetRequest(
		userProfile.OrganizationID,
		createRequestAssetParams,
		ctr.NotificationRepo,
		ctr.UserRepo,
		usersIdGm)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	err = ctr.assetRepository.UpdateAssetUser(createRequestAssetParams.AssetId, 0, createRequestAssetParams.StatusAsset, createRequestAssetParams.StatusReq)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	_, err = ctr.FcmTokenRepo.SelectMultiFcmTokens(usersIdGm, createRequestAssetParams.UserID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	// body = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + body
	// if len(registrationTokens) > 0 {
	// 	for _, token := range registrationTokens {
	// 		if token != "" {
	// 			err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", body, link)
	// 			if err != nil && err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
	// 				"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
	// 				_ = ctr.FcmTokenRepo.DeleteFcmToken(token)
	// 			}
	// 		}
	// 	}
	// }

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create request asset successfully.",
	})
}

func (ctr *Controller) UpdateAsset(c echo.Context) error {
	editAssetParams := new(param.EditAssetParams)

	if err := c.Bind(editAssetParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	_, err := valid.ValidateStruct(editAssetParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.assetRepository.UpdateAsset(userProfile.OrganizationID, editAssetParams)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates") {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Duplicate asset code value violates",
			Data:    err,
		})
	}

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update asset successfully.",
	})
}

func (ctr *Controller) GetAssetByID(c echo.Context) error {
	getAssetParams := new(param.GetAssetParams)

	if err := c.Bind(getAssetParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getAssetParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	asset, err := ctr.assetRepository.SelectAssetByID(getAssetParams.ID)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	dataResponse := map[string]interface{}{
		"asset_type_id":       asset.AssetTypeId,
		"asset_code":          asset.AssetCode,
		"asset_name":          asset.AssetName,
		"user_id":             asset.UserId,
		"managed_by":          asset.ManagedBy,
		"branch_id":           asset.BranchId,
		"status":              asset.Status,
		"description":         asset.Description,
		"purchase_price":      asset.PurchasePrice,
		"date_of_purchase":    asset.DateOfPurchase,
		"license_end_date":    asset.LicenseEndDate,
		"date_started_use":    asset.DateStartedUse,
		"depreciation_period": asset.DepreciationPeriod,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get asset successfully.",
		Data:    dataResponse,
	})
}

func (ctr *Controller) DeleteAssetByID(c echo.Context) error {
	deleteParams := new(param.DeleteParams)

	if err := c.Bind(deleteParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(deleteParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.assetRepository.DeleteAsset(
		deleteParams.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove asset successfully.",
	})
}

func (ctr *Controller) DeleteAssetTypeByID(c echo.Context) error {
	deleteParams := new(param.DeleteParams)

	if err := c.Bind(deleteParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(deleteParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.assetRepository.DeleteAssetType(
		deleteParams.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove asset type successfully.",
	})
}

func (ctr *Controller) CreateAssetLog(c echo.Context) error {
	createAssetLogParams := new(param.CreateAssetLogParams)

	if err := c.Bind(createAssetLogParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createAssetLogParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	now := time.Now()

	if createAssetLogParams.CurrentUserID != 0 {
		_, err := ctr.assetRepository.UpdateAssetLog(userProfile.OrganizationID, createAssetLogParams.CurrentUserID, createAssetLogParams.AssetID, now)
	
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	}

	if createAssetLogParams.NewUserID != 0 {
		err = ctr.assetRepository.CreateAssetLog(userProfile.OrganizationID, createAssetLogParams.AssetID,
			createAssetLogParams.NewUserID, now)

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create new asset log successfully.",
	})
}

func monthsCountSince(createdAtTime time.Time) int {
	now := time.Now()
	months := 0
	month := createdAtTime.Month()
	for createdAtTime.Before(now) {
		createdAtTime = createdAtTime.Add(time.Hour * 24)
		nextMonth := createdAtTime.Month()
		if nextMonth != month {
			months++
		}
		month = nextMonth
	}

	return months
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func (ctr *Controller) DownloadTemplateImportAsset(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)

	f := excelize.NewFile()
	activeSheetName := "Sheet1"
	f.NewSheet(activeSheetName)
	beginRowPos := 7
	beginRowPosStr := strconv.Itoa(beginRowPos)

	// set title
	f.SetCellValue(activeSheetName, "F2", "Import asset")
	titleStyle, _ := f.NewStyle(`{
		"font":{
		"size": 16,
		"bold": true
		},
		"alignment":{
		"horizontal":"center"
		}
	}`)
	f.SetCellStyle(activeSheetName, "F2", "F2", titleStyle)

	// set description
	colorCellDescriptionStyle, _ := f.NewStyle(`{
		"fill":{
		"type":"pattern",
		"color":["#fce4d6"],
		"pattern":1
		}
	}`)
	f.SetCellStyle(activeSheetName, "B4", "B4", colorCellDescriptionStyle)
	f.SetCellValue(activeSheetName, "C4", "Cell have pink color is required")

	f.SetColWidth(activeSheetName, "A", "Q", 20)
	f.SetColWidth(activeSheetName, "E", "E", 40)
	f.SetColWidth(activeSheetName, "G", "G", 40)
	f.SetColWidth(activeSheetName, "K", "K", 60)
	f.SetColWidth(activeSheetName, "N", "Q", 30)

	borderStyleStr := `
        "border": [
            {
                "type": "left",
                "color": "#000000",
                "style": 1
            },
            {
                "type": "top",
                "color": "#000000",
                "style": 1
            },
            {
                "type": "bottom",
                "color": "#000000",
                "style": 1
            },
            {
                "type": "right",
                "color": "#000000",
                "style": 1
            }
        ] `

	headStyle, _ := f.NewStyle(`{
		"font":{
		"size": 14,
		"color": "#FFFFFF"
		},
		"alignment":{
		"horizontal":"center",
		"vertical":"center"
		},
		"fill":{
		"type":"pattern",
		"color":["#00b050"],
		"pattern":1
		},
		` + borderStyleStr + `
	}`)

	fontStyleBodyStr := `"font":{
		"size": 13,
		"color": "#000000"
	}`

	alignmentStyleBodyStr := `"alignment":{
		"horizontal":"left",
		"vertical":"center"
	}`

	fillStyleRequireStr := `"fill":{
		"type":"pattern",
		"color":["#fce4d6"],
		"pattern":1
	}`

	styleNumberFormat := `"number_format":49`

	bodyStyle, _ := f.NewStyle(`{
		` + fontStyleBodyStr + `,
		` + alignmentStyleBodyStr + `,
		` + borderStyleStr + `,
		` + styleNumberFormat + `
	}`)

	bodyRequireStyle, _ := f.NewStyle(`{
		` + fontStyleBodyStr + `,
		` + alignmentStyleBodyStr + `,
		` + borderStyleStr + `,
		` + fillStyleRequireStr + `,
		` + styleNumberFormat + `
	}`)

	var categoriesByLanguage map[string]string
	if userProfile.LanguageId == cf.EnLanguageId {
		categoriesByLanguage = cf.EnAssetCategories
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categoriesByLanguage = cf.VnAssetCategories
	} else {
		categoriesByLanguage = cf.JpAssetCategories
	}

	categories := map[string]string{
		"A" + beginRowPosStr: categoriesByLanguage["Asset type code"],
		"B" + beginRowPosStr: categoriesByLanguage["Asset code"],
		"C" + beginRowPosStr: categoriesByLanguage["Asset name"],
		"D" + beginRowPosStr: categoriesByLanguage["Employee Id"],
		"E" + beginRowPosStr: categoriesByLanguage["User name"],
		"F" + beginRowPosStr: categoriesByLanguage["Managed by Id"],
		"G" + beginRowPosStr: categoriesByLanguage["Managed by name"],
		"H" + beginRowPosStr: categoriesByLanguage["Branch Id"],
		"I" + beginRowPosStr: categoriesByLanguage["Branch"],
		"J" + beginRowPosStr: categoriesByLanguage["Status"],
		"K" + beginRowPosStr: categoriesByLanguage["Description"],
		"L" + beginRowPosStr: categoriesByLanguage["Purchase price"],
		"M" + beginRowPosStr: categoriesByLanguage["Purchase date"],
		"N" + beginRowPosStr: categoriesByLanguage["Warranty expiry date"],
		"O" + beginRowPosStr: categoriesByLanguage["Used since"],
		"P" + beginRowPosStr: categoriesByLanguage["Depreciation end date"],
		"Q" + beginRowPosStr: categoriesByLanguage["Depreciation period"],
	}

	for k, v := range categories {
		_ = f.SetCellValue(activeSheetName, k, v)
	}

	f.SetCellStyle(activeSheetName, "A"+beginRowPosStr, "Q"+beginRowPosStr, headStyle)

	// add 200 row with style for user import
	addSampleRow := 200 + beginRowPos + 1
	f.SetCellStyle(activeSheetName, "A"+strconv.Itoa(beginRowPos+1), "A"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "B"+strconv.Itoa(beginRowPos+1), "B"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "C"+strconv.Itoa(beginRowPos+1), "C"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "D"+strconv.Itoa(beginRowPos+1), "D"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "E"+strconv.Itoa(beginRowPos+1), "E"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "F"+strconv.Itoa(beginRowPos+1), "F"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "G"+strconv.Itoa(beginRowPos+1), "G"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "H"+strconv.Itoa(beginRowPos+1), "H"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "I"+strconv.Itoa(beginRowPos+1), "I"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "J"+strconv.Itoa(beginRowPos+1), "J"+strconv.Itoa(addSampleRow), bodyRequireStyle)
	f.SetCellStyle(activeSheetName, "K"+strconv.Itoa(beginRowPos+1), "K"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "L"+strconv.Itoa(beginRowPos+1), "L"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "M"+strconv.Itoa(beginRowPos+1), "M"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "N"+strconv.Itoa(beginRowPos+1), "N"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "O"+strconv.Itoa(beginRowPos+1), "O"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "P"+strconv.Itoa(beginRowPos+1), "P"+strconv.Itoa(addSampleRow), bodyStyle)
	f.SetCellStyle(activeSheetName, "Q"+strconv.Itoa(beginRowPos+1), "Q"+strconv.Itoa(addSampleRow), bodyStyle)

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}

func (ctr *Controller) ImportExcelAsset(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	rows, errEx := ex.ReadExcelFile(file)
	if errEx != "" {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: errEx,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	userProfileRecords, err := ctr.UserRepo.SelectEmployeeIdByOrganizationId(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	branches, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	assetTypeList, err := ctr.assetRepository.SelectAssetType(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var categoriesByLanguage map[string]string
	if userProfile.LanguageId == cf.EnLanguageId {
		categoriesByLanguage = cf.EnAssetCategories
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categoriesByLanguage = cf.VnAssetCategories
	} else {
		categoriesByLanguage = cf.JpAssetCategories
	}

	var assetArray []m.Asset
	for i, col := range rows {
		index := i + 1
		if i <= 6 {
			continue
		}

		if len(col) == 0 {
			continue
		}

		importAssetObj := m.Asset{}

		if col[0] != "" || col[1] != "" || col[2] != "" {
			// process Asset type
			if col[0] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: asset type can't not be blank"], strconv.Itoa(index)),
				})
			}

			assetTypeCode, err := strconv.Atoi(col[0])

			if err != nil {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: asset type should be number"], strconv.Itoa(index)),
				})
			}

			checkAssetTypeExist := false
			for _, assetTypeItem := range assetTypeList {
				if assetTypeItem.ID == assetTypeCode {
					checkAssetTypeExist = true
					break
				}
			}

			if !checkAssetTypeExist {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: asset type code not exists"], strconv.Itoa(index)),
				})
			}

			importAssetObj.AssetTypeId = assetTypeCode
			// End process Asset type

			// process Asset code
			if col[1] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: asset code can't not be blank"], strconv.Itoa(index)),
				})
			}

			importAssetObj.AssetCode = col[1]
			// End process Asset code

			// Asset process name
			if col[2] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: asset name can't not be blank"], strconv.Itoa(index)),
				})
			}

			importAssetObj.AssetName = col[2]
			// End process Asset name

			// process Employee Id
			if col[3] != "" {
				var userEmployeeObj param.EmployeeIdAndFullName
				for _, itemUser := range userProfileRecords {
					if itemUser.EmployeeId == col[3] {
						userEmployeeObj = itemUser
						break
					}
				}

				if userEmployeeObj.UserId == 0 {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf(categoriesByLanguage["Row %s: employee id is not exist"], strconv.Itoa(index)),
					})
				}

				importAssetObj.UserId = userEmployeeObj.UserId
			}

			// End process Employee Id

			// process Managed by Id
			if col[5] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Managed by Id can't not be blank"], strconv.Itoa(index)),
				})
			}

			var managerEmployeeObj param.EmployeeIdAndFullName
			for _, itemmanager := range userProfileRecords {
				if itemmanager.EmployeeId == col[5] {
					managerEmployeeObj = itemmanager
					break
				}
			}

			if managerEmployeeObj.UserId == 0 {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Manager employee id is not exist"], strconv.Itoa(index)),
				})
			}

			importAssetObj.ManagedBy = managerEmployeeObj.UserId
			// End process Managed by Id

			// process Branch Id
			if col[7] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Branch Id can't not be blank"], strconv.Itoa(index)),
				})
			}

			branchID, err := strconv.Atoi(col[7])

			if err != nil {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Branch Id should be number"], strconv.Itoa(index)),
				})
			}

			var branchObj param.SelectBranchRecords
			for _, itemBranch := range branches {
				if itemBranch.Id == branchID {
					branchObj = itemBranch
					break
				}
			}

			if branchObj.Id == 0 {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Branch Id not exists"], strconv.Itoa(index)),
				})
			}

			importAssetObj.BranchId = branchObj.Id
			// End process Branch Id

			// process Status
			if col[9] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Status can't not be blank"], strconv.Itoa(index)),
				})
			}

			statusID, err := strconv.Atoi(col[9])

			if err != nil {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Status Id should be number"], strconv.Itoa(index)),
				})
			}

			if cf.AssetStatus[statusID] == "" {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: fmt.Sprintf(categoriesByLanguage["Row %s: Status Id not exists"], strconv.Itoa(index)),
				})
			}

			importAssetObj.Status = statusID
			// End process Status

			// Description
			if col[10] != "" {
				importAssetObj.Description = col[10]
			}
			// End description

			// Purchase price
			if col[11] != "" {
				purchasePrice, err := strconv.ParseFloat(col[11], 64)
				if err != nil {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: Purchase price should be number", strconv.Itoa(index)),
					})
				}

				importAssetObj.PurchasePrice = purchasePrice
			}
			// End purchase price

			// Purchase date
			if col[12] != "" {
				dateOfPurchase, err := time.Parse(cf.FormatDateDatabase, col[12])
				if err != nil {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: Purchase date is invalid, it 's should be like 2006-01-02", strconv.Itoa(index)),
					})
				}

				importAssetObj.DateOfPurchase = dateOfPurchase
			}
			// End purchase date

			// Warranty expiry date
			if col[13] != "" {
				licenseEndDate, err := time.Parse(cf.FormatDateDatabase, col[13])

				if err != nil {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: Warranty expiry date is invalid, it 's should be like 2006-01-02", strconv.Itoa(index)),
					})
				}

				importAssetObj.LicenseEndDate = licenseEndDate
			}
			// End warranty expiry date

			// Used since
			if col[14] != "" {
				usedSince, err := time.Parse(cf.FormatDateDatabase, col[14])

				if err != nil {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: Used since date is invalid, it 's should be like 2006-01-02", strconv.Itoa(i)),
					})
				}

				importAssetObj.DateStartedUse = usedSince
			}
			// End used since

			// Depreciation period
			if col[16] != "" {
				depreciationPeriod, err := strconv.Atoi(col[16])

				if err != nil {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: depreciation period should be number", strconv.Itoa(i)),
					})
				}

				importAssetObj.DepreciationPeriod = depreciationPeriod
			}
			// End depreciation period

			importAssetObj.OrganizationId = userProfile.OrganizationID

			for _, assetItem := range assetArray {
				if assetItem.AssetCode == importAssetObj.AssetCode {
					return c.JSON(http.StatusBadRequest, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: fmt.Sprintf("Row %s: asset is duplicated", strconv.Itoa(i)),
					})
				}
			}

			assetArray = append(assetArray, importAssetObj)
		}
	}

	if len(assetArray) > 0 {
		err = ctr.assetRepository.ImportAsset(userProfile.OrganizationID, assetArray)

		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Import asset successfully.",
	})
}
