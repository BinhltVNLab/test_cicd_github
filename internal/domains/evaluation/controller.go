package evaluation

import (
	"encoding/json"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"net/http"
	"strconv"
	"sync"
)

// TargetEvalController : Controler
type TargetEvalController struct {
	cm.BaseController

	TargetEvaluationRepo rp.EvaluationRepository
	UserRepo             rp.UserRepository
	ProjectRepo          rp.ProjectRepository
	UserProjectRepo      rp.UserProjectRepository
	BranchRepo           rp.BranchRepository

	Cloud gc.StorageUtility
}

// NewTargetEvaluationController : Create Controller
func NewTargetEvaluationController(
	logger echo.Logger,
	targetEvaluationRepo rp.EvaluationRepository,
	userRepo rp.UserRepository,
	projectRepo rp.ProjectRepository,
	userProjectRepo rp.UserProjectRepository,
	branchRepo rp.BranchRepository,
	cloud gc.StorageUtility,
) (ctr *TargetEvalController) {
	ctr = &TargetEvalController{cm.BaseController{}, targetEvaluationRepo, userRepo, projectRepo, userProjectRepo, branchRepo, cloud}
	ctr.Init(logger)
	return
}

// CreateEvaluationForm : User create target form
// Params     : echo.Context
// Returns    : return data with struct JsonResponse
func (ctr *TargetEvalController) CreateEvaluationForm(c echo.Context) error {
	createEvaluationParams := new(param.CreateEvaluationParams)

	if err := c.Bind(createEvaluationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createEvaluationParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	createEvaluationParams.UserID = userProfile.UserProfile.UserID
	createEvaluationParams.OrganizationID = userProfile.OrganizationID
	createEvaluationParams.Status = cf.EvaluationCreatedStatus
	createEvaluationParams.UpdatedBy = userProfile.UserProfile.UserID

	count, err := ctr.TargetEvaluationRepo.CheckExist(
		userProfile.OrganizationID,
		userProfile.UserProfile.UserID,
		createEvaluationParams.Year,
		createEvaluationParams.Quarter,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusNotAcceptable, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Evaluation is already exist",
		})
	}

	evaluationForm, err := ctr.TargetEvaluationRepo.InsertEvaluation(createEvaluationParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	dataResponse := map[string]interface{}{
		"id":              evaluationForm.ID,
		"user_id":         evaluationForm.UserID,
		"content":         evaluationForm.Content,
		"organization_id": evaluationForm.OrganizationID,
		"status":          evaluationForm.Status,
		"quarter":         evaluationForm.Quarter,
		"year":            evaluationForm.Year,
		"updated_by":      evaluationForm.UpdatedBy,
		"created_at":      evaluationForm.CreatedAt,
		"updated_at":      evaluationForm.UpdatedAt,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Form created successfully.",
		Data:    dataResponse,
	})
}

// CheckEvalExist : check eval existed
// Params     : echo.Context
// Returns    : return data with struct JsonResponse
func (ctr *TargetEvalController) CheckEvalExist(c echo.Context) error {
	checkEvalExistParams := new(param.CheckEvalExistParams)
	userProfile := c.Get("user_profile").(m.User)

	if err := c.Bind(checkEvalExistParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}
	evalObj, err := ctr.TargetEvaluationRepo.CheckEvaluationExists(
		userProfile.OrganizationID,
		userProfile.UserProfile.UserID,
		checkEvalExistParams.Year,
		checkEvalExistParams.Quarter,
	)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	dataResponse := map[string]interface{}{
		"id":            evalObj.ID,
		"user_id":       evalObj.UserID,
		"status":        evalObj.Status,
		"content":       evalObj.Content,
		"updated_by":    evalObj.UpdatedBy,
		"quarter":       evalObj.Quarter,
		"year":          evalObj.Year,
		"jp_lang_level": cf.JpLanguageLevel,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Evaluation existed.",
		Data:    dataResponse,
	})
}

// DuplicateEvaluationForm : User duplicate target form
// Params     : echo.Context
// Returns    : return data with struct JsonResponse
func (ctr *TargetEvalController) DuplicateEvaluationForm(c echo.Context) error {
	duplicateEvaluationParams := new(param.DuplicateEvaluationParams)

	if err := c.Bind(duplicateEvaluationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(duplicateEvaluationParams)
	if err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	evalObj, err := ctr.TargetEvaluationRepo.GetEvaluation(duplicateEvaluationParams.EvalFormID, userProfile.OrganizationID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if evalObj.UserID != userProfile.UserProfile.UserID && userProfile.RoleID == cf.UserRoleID {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Not permission to duplicate",
		})
	}

	createEvaluationParams := new(param.CreateEvaluationParams)
	createEvaluationParams.UserID = userProfile.UserProfile.UserID
	createEvaluationParams.OrganizationID = evalObj.OrganizationID

	quarter, year := utils.GetCurrentQuarterAndYear()

	createEvaluationParams.Content = evalObj.Content
	createEvaluationParams.Status = cf.EvaluationCreatedStatus
	createEvaluationParams.Quarter = quarter
	createEvaluationParams.Year = year
	createEvaluationParams.UpdatedBy = userProfile.UserProfile.UserID

	dataResponse := map[string]interface{}{
		"id":            evalObj.ID,
		"user_id":       evalObj.UserID,
		"status":        evalObj.Status,
		"content":       evalObj.Content,
		"updated_by":    evalObj.UpdatedBy,
		"quarter":       evalObj.Quarter,
		"year":          evalObj.Year,
		"jp_lang_level": cf.JpLanguageLevel,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status: cf.SuccessResponseCode,
		Data:   dataResponse,
	})
}

// GetEvaluationForm : Get evaluation form
// Params     : echo.Context
// Returns    : return data with struct JsonResponse
func (ctr *TargetEvalController) GetEvaluationForm(c echo.Context) error {
	getEvaluationParams := new(param.GetEvaluationParams)
	if err := c.Bind(getEvaluationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(getEvaluationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	evalObj, err := ctr.TargetEvaluationRepo.GetEvaluation(getEvaluationParams.EvalFormID, userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if userProfile.RoleID != cf.GeneralManagerRoleID &&
		evalObj.UserID != userProfile.UserProfile.UserID {
		var managedUserIds []int
		records, err := ctr.UserProjectRepo.SelectUserIdsManagedByManager(userProfile.OrganizationID, userProfile.UserProfile.UserID)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		for _, record := range records {
			managedUserIds = append(managedUserIds, record.UserId)
		}

		if !utils.FindIntInSlice(managedUserIds, evalObj.UserID) {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You do not have permission to view.",
			})
		}
	}

	dataResponse := map[string]interface{}{
		"id":            evalObj.ID,
		"user_id":       evalObj.UserID,
		"status":        evalObj.Status,
		"content":       evalObj.Content,
		"updated_by":    evalObj.UpdatedBy,
		"quarter":       evalObj.Quarter,
		"year":          evalObj.Year,
		"jp_lang_level": cf.JpLanguageLevel,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get successfully.",
		Data:    dataResponse,
	})
}

// UpdateEvaluationForm : update evaluation form
// Params     : echo.Context
// Returns    : return data with struct JsonResponse
func (ctr *TargetEvalController) UpdateEvaluationForm(c echo.Context) error {
	updateEvaluationParams := new(param.UpdateEvaluationParams)
	if err := c.Bind(updateEvaluationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(updateEvaluationParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	evalObj, err := ctr.TargetEvaluationRepo.GetEvaluation(updateEvaluationParams.EvalFormID, userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if userProfile.RoleID != cf.GeneralManagerRoleID &&
		evalObj.UserID != userProfile.UserProfile.UserID {
		var managedUserIds []int
		records, err := ctr.UserProjectRepo.SelectUserIdsManagedByManager(userProfile.OrganizationID, userProfile.UserProfile.UserID)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		for _, record := range records {
			managedUserIds = append(managedUserIds, record.UserId)
		}

		if !utils.FindIntInSlice(managedUserIds, evalObj.UserID) {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You do not have permission to update.",
			})
		}
	}

	updateEvaluationParams.UpdatedBy = userProfile.UserProfile.UserID

	if userProfile.RoleID == cf.UserRoleID {
		if updateEvaluationParams.Status != cf.EvaluationCreatedStatus &&
			updateEvaluationParams.Status != cf.EvaluationMemberIsCreatingStatus &&
			updateEvaluationParams.Status != cf.EvaluationMemberEditedStatus {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
				Data:    err,
			})
		}
	}

	err = ctr.TargetEvaluationRepo.UpdateEvaluation(userProfile.OrganizationID, updateEvaluationParams)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Update Failed",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update Successfully",
	})
}

// SearchEvaluationList : get list of evaluation by Organization
// Params : echo.Context
// Returns : return error
func (ctr *TargetEvalController) SearchEvaluationList(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	searchEvaluationListParams := new(param.SearchEvaluationListParams)
	searchEvaluationListParams.RowPerPage = 20

	if err := c.Bind(searchEvaluationListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(searchEvaluationListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	evaluation, totalRow, err := ctr.TargetEvaluationRepo.GetEvaluationList(userProfile.OrganizationID, searchEvaluationListParams)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get Evaluation List Failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	if searchEvaluationListParams.RowPerPage == 0 {
		searchEvaluationListParams.CurrentPage = 1
		searchEvaluationListParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": searchEvaluationListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": searchEvaluationListParams.RowPerPage,
	}

	listEvaluationResponse := []map[string]interface{}{}
	for i := 0; i < len(evaluation); i++ {
		var base64Img []byte = nil
		if evaluation[i].Avatar != "" {
			base64Img, err = ctr.Cloud.GetFileByFileName(evaluation[i].Avatar, cf.AvatarFolderGCS)

			if err != nil {
				ctr.Logger.Error(err)
				base64Img = nil
			}
		}
		itemDataResponse := map[string]interface{}{
			"id":              evaluation[i].ID,
			"name":            evaluation[i].Name,
			"updated_by_name": evaluation[i].UpdatedByName,
			"quarter":         evaluation[i].Quarter,
			"year":            evaluation[i].Year,
			"branch":          evaluation[i].Branch,
			"rank":            evaluation[i].Rank,
			"status":          evaluation[i].Status,
			"last_updated":    evaluation[i].UpdatedAt.Format(cf.FormatDateDisplay),
			"updated_by":      evaluation[i].UpdatedBy,
			"avatar":          base64Img,
		}

		listEvaluationResponse = append(listEvaluationResponse, itemDataResponse)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	for _, record := range branchRecords {
		branches[record.Id] = record.Name
	}

	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userList := make(map[int]string)
	for i := 0; i < len(users); i++ {
		userList[users[i].UserID] = users[i].FullName
	}

	projects, err := ctr.ProjectRepo.SelectProjectsByOrganizationId(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	projectList := make(map[int]string)
	if len(projects) > 0 {
		for i := 0; i < len(projects); i++ {
			projectList[projects[i].ID] = projects[i].Name
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination":         pagination,
			"branch_select_box":  branches,
			"rank_select_box":    cf.EvaluationRankList,
			"quarter_select_box": cf.QuarterList,
			"status_select_box":  cf.EvaluationStatus,
			"evaluations":        listEvaluationResponse,
			"users":              userList,
			"projects":           projectList,
		},
	})
}

// EvaluationListByUserID : get list of evaluation by Organization
// Params : echo.Context
// Returns : return error
func (ctr *TargetEvalController) EvaluationListByUserID(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	evaluation := ctr.TargetEvaluationRepo.GetEvaluationListByUserID(userProfile.OrganizationID, userProfile.UserProfile.UserID)

	listEvaluationResponse := []map[string]interface{}{}
	for i := 0; i < len(evaluation); i++ {
		itemDataResponse := map[string]interface{}{
			"id":      evaluation[i].ID,
			"quarter": evaluation[i].Quarter,
			"year":    evaluation[i].Year,
		}

		listEvaluationResponse = append(listEvaluationResponse, itemDataResponse)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    listEvaluationResponse,
	})
}

// DeleteEvaluation : delete Evaluation by id
// Params : echo.Context
// Returns : object
func (ctr *TargetEvalController) DeleteEvaluation(c echo.Context) error {
	deleteEvaluationParams := new(param.DeleteEvaluationParams)
	if err := c.Bind(deleteEvaluationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(deleteEvaluationParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	err := ctr.TargetEvaluationRepo.DeleteEvaluation(deleteEvaluationParams.EvaluationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Deleted",
	})
}

func (ctr *TargetEvalController) GetCommentTwoConsecutiveQuarter(c echo.Context) error {
	params := new(param.GetCommentTwoConsecutiveQuarterParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	if (params.LastYear == params.Year && params.LastQuarter > params.Quarter) || params.LastYear > params.Year {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.TargetEvaluationRepo.SelectCommentTwoConsecutiveQuarter(userProfile.OrganizationID, params)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.SuccessResponseCode,
				Message: "Get comment two consecutive quarter successful",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var dataResponse []map[string]interface{}
	for _, record := range records {
		data := map[string]interface{}{
			"user_id":      record.UserID,
			"full_name":    record.FullName,
			"rank":         cf.EvaluationRankList[record.Rank],
			"score":        record.Score,
			"comment":      record.Comment,
			"last_rank":    cf.EvaluationRankList[record.LastRank],
			"last_score":   record.LastScore,
			"last_comment": record.LastComment,
		}

		var base64Img []byte
		if record.Avatar != "" {
			base64Img, err = ctr.Cloud.GetFileByFileName(record.Avatar, cf.AvatarFolderGCS)
			if err != nil {
				ctr.Logger.Error(err)
			}
		}
		data["avatar"] = base64Img

		dataResponse = append(dataResponse, data)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get comment two consecutive quarter successful",
		Data:    dataResponse,
	})
}

// ExportExcel : Export evaluation to excel
func (ctr *TargetEvalController) ExportMultipleExcel(c echo.Context) error {
	evaluationIdsStr := c.FormValue("evaluation_ids")
	if evaluationIdsStr == "" {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	var evaluationIds []int
	err := json.Unmarshal([]byte(evaluationIdsStr), &evaluationIds)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	done := make(chan map[string]interface{}, len(evaluationIds))
	var wg sync.WaitGroup
	for _, evaluationId := range evaluationIds {
		wg.Add(1)
		go func(evaluationId int) {
			buf, fileName := ctr.exportExcel(userProfile, evaluationId)
			data := map[string]interface{}{
				"buf":       buf,
				"file_name": fileName,
			}
			done <- data
		}(evaluationId)
	}

	dataResponse := make([]map[string]interface{}, 0)
	go func() {
		for t := range done {
			dataResponse = append(dataResponse, t)
			wg.Done()
		}
	}()
	wg.Wait()

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Export excel successful",
		Data:    dataResponse,
	})
}

func (ctr *TargetEvalController) exportExcel(userProfile m.User, evaluationId int) ([]byte, string) {
	var categories map[string]string
	if userProfile.LanguageId == cf.EnLanguageId {
		categories = cf.EvaluationCategoriesEN
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categories = cf.EvaluationCategoriesVN
	} else {
		categories = cf.EvaluationCategoriesJP
	}

	evalObj, err := ctr.TargetEvaluationRepo.GetEvaluationToExport(evaluationId, userProfile.OrganizationID)
	if err != nil {
		panic(err)
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "N", 15)
	_ = f.SetColWidth("Sheet1", "I", "J", 15)
	_ = f.SetColWidth("Sheet1", "L", "M", 17)

	leftBorderBold, _ := f.NewStyle(cf.ExcelStyles["leftBorderBold"])
	topBorderBold, _ := f.NewStyle(cf.ExcelStyles["topBorderBold"])
	bottomBorderBold, _ := f.NewStyle(cf.ExcelStyles["bottomBorderBold"])
	rightBorderBold, _ := f.NewStyle(cf.ExcelStyles["rightBorderBold"])
	fillBackgroundColor, _ := f.NewStyle(cf.ExcelStyles["fillBackgroundColor"])

	topAndRightBorderThin, _ := f.NewStyle(cf.ExcelStyles["topAndRightBorderThin"])
	rightAndLeftBorderBold, _ := f.NewStyle(cf.ExcelStyles["rightAndLeftBorderBold"])
	topAndBottomBorderBold, _ := f.NewStyle(cf.ExcelStyles["topAndBottomBorderBold"])
	leftBorderThin, _ := f.NewStyle(cf.ExcelStyles["leftBorderThin"])
	topAndLeftBorderThin, _ := f.NewStyle(cf.ExcelStyles["topAndLeftBorderThin"])
	topDotLeftContinuousBorder, _ := f.NewStyle(cf.ExcelStyles["topDotLeftContinuousBorder"])
	allThinBorderCenter, _ := f.NewStyle(cf.ExcelStyles["allThinBorderCenter"])
	rightAndBottomBorderThinFontBold, _ := f.NewStyle(cf.ExcelStyles["rightAndBottomBorderThinFontBold"])
	topAndRightBorderThinFontBold, _ := f.NewStyle(cf.ExcelStyles["topAndRightBorderThinFontBold"])
	topBoldRightBottomBorderThinFontBold, _ := f.NewStyle(cf.ExcelStyles["topBoldRightBottomBorderThinFontBold"])
	topDotBorder, _ := f.NewStyle(cf.ExcelStyles["topDotBorder"])

	allBoldBorderUser, _ := f.NewStyle(cf.ExcelStyles["allBoldBorderUser"])
	bottomDotTopAndRightBorderThinUser, _ := f.NewStyle(cf.ExcelStyles["bottomDotTopAndRightBorderThinUser"])
	allThinBorderUser, _ := f.NewStyle(cf.ExcelStyles["allThinBorderUser"])
	rightAndBottomBorderThinUser, _ := f.NewStyle(cf.ExcelStyles["rightAndBottomBorderThinUser"])
	leftAndBottomBorderThinFontBold, _ := f.NewStyle(cf.ExcelStyles["leftAndBottomBorderThinFontBold"])

	allBoldBorderSupervisor, _ := f.NewStyle(cf.ExcelStyles["allBoldBorderSupervisor"])
	topDotRightContinuousBorderUser, _ := f.NewStyle(cf.ExcelStyles["topDotRightContinuousBorderUser"])
	bottomDotTopAndRightHorizontalLeftBorderThinUser, _ := f.NewStyle(cf.ExcelStyles["bottomDotTopAndRightHorizontalLeftBorderThinUser"])
	topBottomDotRightContinuousHorizontalLeftBorderUser, _ := f.NewStyle(cf.ExcelStyles["topBottomDotRightContinuousHorizontalLeftBorderUser"])
	topLeftBoldRightBottomBorderThinFontBold, _ := f.NewStyle(cf.ExcelStyles["topLeftBoldRightBottomBorderThinFontBold"])
	allThinBorderSupervisor, _ := f.NewStyle(cf.ExcelStyles["allThinBorderSupervisor"])
	topDotRightContinuousBorderSupervisor, _ := f.NewStyle(cf.ExcelStyles["topDotRightContinuousBorderSupervisor"])
	topAndRightBorderThinSupervisor, _ := f.NewStyle(cf.ExcelStyles["topAndRightBorderThinSupervisor"])
	topRightBoldLeftBottomThinBorderFontBold, _ := f.NewStyle(cf.ExcelStyles["topRightBoldLeftBottomThinBorderFontBold"])

	firstCategories := map[string]string{
		"A1": categories["CT1"],
		"I1": categories["CT4"],
		"I2": categories["CT5"],

		"B3": categories["CT6"],
		"D3": categories["CT7"],
		"E3": categories["CT8"],
		"F3": categories["CT9"],
		"I3": categories["CT10"],
		"J3": categories["CT11"],
		"L3": categories["CT12"],
		"M3": categories["CT13"],
		"N3": categories["CT14"],
		"Q3": categories["CT15"],

		"B4": categories["CT16"],
		"Q4": categories["CT17"],
		"B6": categories["CT18"],
		"I6": categories["CT19"],

		"B8":  categories["CT20"],
		"B9":  categories["CT21"],
		"B10": categories["CT22"],
		"C10": categories["CT23"],
		"F10": categories["CT24"],
		"J10": categories["CT25"],
		"M10": categories["CT26"],
		"N10": categories["CT27"],

		"B16": categories["CT28"],
		"B17": categories["CT29"],
		"B18": categories["CT30"],
		"C18": categories["CT31"],
		"F18": categories["CT32"],
		"J18": categories["CT33"],
		"M18": categories["CT34"],
		"N18": categories["CT35"],
	}

	if userProfile.LanguageId == cf.JpLanguageId {
		firstCategories["D1"] = categories["CT2"]
		firstCategories["G1"] = categories["CT3"]
	} else {
		firstCategories["C1"] = categories["CT2"]
		firstCategories["F1"] = categories["CT3"]
	}
	for k, v := range firstCategories {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(evalObj.OrganizationID)
	if err != nil {
		panic(err)
	}

	var branch string
	if len(branchRecords) > 0 {
		for _, record := range branchRecords {
			if evalObj.Branch == record.Id {
				branch = record.Name
			}
		}
	}

	values := map[string]interface{}{
		"C3": evalObj.FullName,
		"C4": branch,
		"I4": evalObj.Content.Result.Points,
		"J4": cf.EvaluationRankList[evalObj.Content.Result.Rank],
		"F4": evalObj.Role,
		"J6": 100,
	}

	if userProfile.LanguageId == cf.JpLanguageId {
		values["C1"] = evalObj.Year
		values["F1"] = evalObj.Quarter
	} else {
		values["D1"] = evalObj.Year
		values["G1"] = evalObj.Quarter
	}

	for k, v := range values {
		_ = f.SetCellValue("Sheet1", k, v)
	}
	_ = f.MergeCell("Sheet1", "B10", "B11")
	_ = f.MergeCell("Sheet1", "C10", "E11")
	_ = f.MergeCell("Sheet1", "F10", "I11")
	_ = f.MergeCell("Sheet1", "J10", "L11")
	_ = f.MergeCell("Sheet1", "M10", "M11")
	_ = f.MergeCell("Sheet1", "N10", "N11")

	_ = f.MergeCell("Sheet1", "C12", "E12")
	_ = f.MergeCell("Sheet1", "F12", "I12")
	_ = f.MergeCell("Sheet1", "J12", "L12")

	_ = f.MergeCell("Sheet1", "C13", "E13")
	_ = f.MergeCell("Sheet1", "F13", "I13")
	_ = f.MergeCell("Sheet1", "J13", "L13")

	_ = f.MergeCell("Sheet1", "C14", "E14")
	_ = f.MergeCell("Sheet1", "F14", "I14")
	_ = f.MergeCell("Sheet1", "J14", "L14")

	_ = f.MergeCell("Sheet1", "B18", "B19")
	_ = f.MergeCell("Sheet1", "C18", "E19")
	_ = f.MergeCell("Sheet1", "F18", "I19")
	_ = f.MergeCell("Sheet1", "J18", "L19")
	_ = f.MergeCell("Sheet1", "M18", "M19")
	_ = f.MergeCell("Sheet1", "N18", "N19")

	_ = f.SetColStyle("Sheet1", "A:T", fillBackgroundColor)
	if userProfile.LanguageId == cf.JpLanguageId {
		_ = f.SetCellStyle("Sheet1", "C1", "C1", allThinBorderCenter)
		_ = f.SetCellStyle("Sheet1", "F1", "F1", allThinBorderCenter)
	} else {
		_ = f.SetCellStyle("Sheet1", "D1", "D1", allThinBorderCenter)
		_ = f.SetCellStyle("Sheet1", "G1", "G1", allThinBorderCenter)
	}

	_ = f.SetCellStyle("Sheet1", "B2", "F2", bottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "B5", "F5", topBorderBold)
	_ = f.SetCellStyle("Sheet1", "A3", "A4", rightBorderBold)
	_ = f.SetCellStyle("Sheet1", "G3", "G4", leftBorderBold)
	_ = f.SetCellStyle("Sheet1", "B3", "B3", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "C3", "C3", rightAndBottomBorderThinUser)
	_ = f.SetCellStyle("Sheet1", "D3", "E3", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "F3", "F3", leftAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "B4", "B4", topAndRightBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "C4", "E4", bottomDotTopAndRightBorderThinUser)
	_ = f.SetCellStyle("Sheet1", "F4", "F4", topAndLeftBorderThin)

	_ = f.SetCellStyle("Sheet1", "I2", "J2", bottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "I5", "J5", topAndBottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "H3", "H4", rightBorderBold)
	_ = f.SetCellStyle("Sheet1", "K3", "K4", rightAndLeftBorderBold)
	_ = f.SetCellStyle("Sheet1", "I3", "I3", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "J3", "J3", leftAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "I4", "I4", topAndRightBorderThin)
	_ = f.SetCellStyle("Sheet1", "J4", "J4", topAndLeftBorderThin)

	_ = f.SetCellStyle("Sheet1", "L2", "N2", bottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "L5", "N5", topBorderBold)
	_ = f.SetCellStyle("Sheet1", "O3", "O4", leftBorderBold)
	_ = f.SetCellStyle("Sheet1", "L3", "M3", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "N3", "N3", leftAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "L4", "M4", topAndRightBorderThinSupervisor)
	_ = f.SetCellStyle("Sheet1", "N4", "N4", topAndLeftBorderThin)

	_ = f.SetCellStyle("Sheet1", "P3", "P3", allBoldBorderUser)
	_ = f.SetCellStyle("Sheet1", "P4", "P4", allBoldBorderSupervisor)
	_ = f.SetCellStyle("Sheet1", "Q3", "Q3", fillBackgroundColor)

	_ = f.SetCellStyle("Sheet1", "I7", "J7", topBorderBold)
	_ = f.SetCellStyle("Sheet1", "H6", "H6", rightBorderBold)
	_ = f.SetCellStyle("Sheet1", "K6", "K6", leftBorderBold)
	_ = f.SetCellStyle("Sheet1", "J6", "J6", leftBorderThin)

	commonValue := map[string]interface{}{
		"B12": evalObj.Content.Common.Weight,
		"C12": evalObj.Content.Common.Value,
		"F12": evalObj.Content.Common.Numeric,
		"J12": evalObj.Content.Common.ActualEval,
		"M12": evalObj.Content.Common.CompletionRate,
		"N12": evalObj.Content.Common.Points,
	}
	for k, v := range commonValue {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	_ = f.SetCellStyle("Sheet1", "B9", "N9", bottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "B15", "N15", topBorderBold)
	_ = f.SetCellStyle("Sheet1", "A10", "A14", rightBorderBold)
	_ = f.SetCellStyle("Sheet1", "O10", "O14", leftBorderBold)
	_ = f.SetCellStyle("Sheet1", "B10", "M11", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "N10", "N11", leftAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "B12", "B12", bottomDotTopAndRightBorderThinUser)
	_ = f.SetCellStyle("Sheet1", "C12", "E12", bottomDotTopAndRightHorizontalLeftBorderThinUser)
	_ = f.SetCellStyle("Sheet1", "F12", "L12", bottomDotTopAndRightBorderThinUser)
	_ = f.SetCellStyle("Sheet1", "M12", "M12", topAndRightBorderThin)
	_ = f.SetCellStyle("Sheet1", "N12", "N12", topAndLeftBorderThin)
	_ = f.SetCellStyle("Sheet1", "B13", "B14", topDotRightContinuousBorderUser)
	_ = f.SetCellStyle("Sheet1", "C13", "E14", topBottomDotRightContinuousHorizontalLeftBorderUser)
	_ = f.SetCellStyle("Sheet1", "F13", "L14", topDotRightContinuousBorderUser)
	_ = f.SetCellStyle("Sheet1", "M13", "N14", topDotLeftContinuousBorder)

	_ = f.SetCellStyle("Sheet1", "B17", "N17", bottomBorderBold)
	_ = f.SetCellStyle("Sheet1", "A18", "A19", rightBorderBold)
	_ = f.SetCellStyle("Sheet1", "O18", "O19", leftBorderBold)
	_ = f.SetCellStyle("Sheet1", "B18", "M19", rightAndBottomBorderThinFontBold)
	_ = f.SetCellStyle("Sheet1", "N18", "N19", leftAndBottomBorderThinFontBold)

	_ = f.SetRowHeight("Sheet1", 12, 40)
	_ = f.SetRowHeight("Sheet1", 13, 40)
	_ = f.SetRowHeight("Sheet1", 14, 40)

	idx := 20
	for i, personal := range evalObj.Content.Individual {
		pos := idx + i
		personalValue := map[string]interface{}{
			"B" + strconv.Itoa(pos): personal.Weight,
			"C" + strconv.Itoa(pos): personal.Item,
			"F" + strconv.Itoa(pos): personal.Goal,
			"J" + strconv.Itoa(pos): personal.ActualEval,
			"M" + strconv.Itoa(pos): personal.CompletionRate,
			"N" + strconv.Itoa(pos): personal.Points,
		}

		for k, v := range personalValue {
			_ = f.SetCellValue("Sheet1", k, v)
		}

		_ = f.MergeCell("Sheet1", "C"+strconv.Itoa(pos), "E"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "F"+strconv.Itoa(pos), "I"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos))

		if len(evalObj.Content.Individual) == 0 || len(evalObj.Content.Individual) == 1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "E"+strconv.Itoa(pos), bottomDotTopAndRightHorizontalLeftBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "F"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topAndLeftBorderThin)
		} else {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "E"+strconv.Itoa(pos), topBottomDotRightContinuousHorizontalLeftBorderUser)
			_ = f.SetCellStyle("Sheet1", "F"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topDotLeftContinuousBorder)
		}

		_ = f.SetCellStyle("Sheet1", "A"+strconv.Itoa(pos), "A"+strconv.Itoa(pos), rightBorderBold)
		_ = f.SetCellStyle("Sheet1", "O"+strconv.Itoa(pos), "O"+strconv.Itoa(pos), leftBorderBold)

		if len(evalObj.Content.Individual) == 0 || len(evalObj.Content.Individual) == 1 || i == len(evalObj.Content.Individual)-1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos+1), "B"+strconv.Itoa(pos+2), topLeftBoldRightBottomBorderThinFontBold)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos+1), "M"+strconv.Itoa(pos+2), topBoldRightBottomBorderThinFontBold)
			_ = f.SetCellStyle("Sheet1", "N"+strconv.Itoa(pos+1), "N"+strconv.Itoa(pos+2), topRightBoldLeftBottomThinBorderFontBold)
		}

		_ = f.SetRowHeight("Sheet1", pos, 40)
	}

	idx += len(evalObj.Content.Individual)
	_ = f.MergeCell("Sheet1", "B"+strconv.Itoa(idx), "B"+strconv.Itoa(idx+1))
	_ = f.MergeCell("Sheet1", "C"+strconv.Itoa(idx), "E"+strconv.Itoa(idx+1))
	_ = f.MergeCell("Sheet1", "F"+strconv.Itoa(idx), "I"+strconv.Itoa(idx+1))
	_ = f.MergeCell("Sheet1", "J"+strconv.Itoa(idx), "L"+strconv.Itoa(idx+1))
	_ = f.MergeCell("Sheet1", "M"+strconv.Itoa(idx), "M"+strconv.Itoa(idx+1))
	_ = f.MergeCell("Sheet1", "N"+strconv.Itoa(idx), "N"+strconv.Itoa(idx+1))

	idx += 1
	secondCategories := map[string]string{
		"B" + strconv.Itoa(idx): categories["CT36"],
		"C" + strconv.Itoa(idx): categories["CT37"],
		"F" + strconv.Itoa(idx): categories["CT38"],
		"J" + strconv.Itoa(idx): categories["CT39"],
		"M" + strconv.Itoa(idx): categories["CT40"],
		"N" + strconv.Itoa(idx): categories["CT41"],
	}
	for k, v := range secondCategories {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	idx += 1
	count := 0
	for i, project := range evalObj.Content.Projects {
		if project.Weight == 0 {
			continue
		}
		prj, err := ctr.ProjectRepo.GetProjectByID(project.ID)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			panic(err)
		}

		var targetName string
		for _, target := range prj.Targets {
			if target.Quarter == evalObj.Quarter && target.Year == evalObj.Year {
				targetName = target.TargetContent
			}
		}

		pos := idx + count
		_ = f.MergeCell("Sheet1", "C"+strconv.Itoa(pos), "E"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "F"+strconv.Itoa(pos), "I"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos))

		projectValue := map[string]interface{}{
			"B" + strconv.Itoa(pos): project.Weight,
			"C" + strconv.Itoa(pos): prj.Name,
			"F" + strconv.Itoa(pos): targetName,
			"J" + strconv.Itoa(pos): project.SelfAssessment,
			"M" + strconv.Itoa(pos): project.SuperiorEval,
			"N" + strconv.Itoa(pos): project.Points,
		}

		for k, v := range projectValue {
			_ = f.SetCellValue("Sheet1", k, v)
		}

		if len(evalObj.Content.Projects) == 0 || len(evalObj.Content.Projects) == 1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "I"+strconv.Itoa(pos), bottomDotTopAndRightHorizontalLeftBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "M"+strconv.Itoa(pos), topAndRightBorderThinSupervisor)
			_ = f.SetCellStyle("Sheet1", "N"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topAndLeftBorderThin)
		} else {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "I"+strconv.Itoa(pos), topBottomDotRightContinuousHorizontalLeftBorderUser)
			_ = f.SetCellStyle("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "M"+strconv.Itoa(pos), topDotRightContinuousBorderSupervisor)
			_ = f.SetCellStyle("Sheet1", "N"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topDotLeftContinuousBorder)
		}

		_ = f.SetCellStyle("Sheet1", "A"+strconv.Itoa(pos), "A"+strconv.Itoa(pos), rightBorderBold)
		_ = f.SetCellStyle("Sheet1", "O"+strconv.Itoa(pos), "O"+strconv.Itoa(pos), leftBorderBold)

		if len(evalObj.Content.Projects) == 0 || len(evalObj.Content.Projects) == 1 || i == len(evalObj.Content.Projects)-1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos+1), "N"+strconv.Itoa(pos+1), topDotBorder)
		}

		_ = f.SetRowHeight("Sheet1", pos, 100)
		count++
	}

	idx += count
	for i, challenge := range evalObj.Content.Challenges {
		pos := idx + i
		otherValue := map[string]interface{}{
			"B" + strconv.Itoa(pos): challenge.Weight,
			"C" + strconv.Itoa(pos): challenge.Name,
			"F" + strconv.Itoa(pos): challenge.Actions,
			"J" + strconv.Itoa(pos): challenge.SelfAssessment,
			"M" + strconv.Itoa(pos): challenge.SuperiorEval,
			"N" + strconv.Itoa(pos): challenge.Points,
		}

		for k, v := range otherValue {
			_ = f.SetCellValue("Sheet1", k, v)
		}

		_ = f.MergeCell("Sheet1", "C"+strconv.Itoa(pos), "E"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "F"+strconv.Itoa(pos), "I"+strconv.Itoa(pos))
		_ = f.MergeCell("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos))

		if len(evalObj.Content.Challenges) == 0 || len(evalObj.Content.Challenges) == 1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "I"+strconv.Itoa(pos), bottomDotTopAndRightHorizontalLeftBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), bottomDotTopAndRightBorderThinUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "M"+strconv.Itoa(pos), topAndRightBorderThinSupervisor)
			_ = f.SetCellStyle("Sheet1", "N"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topAndLeftBorderThin)
		} else {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos), "B"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "C"+strconv.Itoa(pos), "I"+strconv.Itoa(pos), topBottomDotRightContinuousHorizontalLeftBorderUser)
			_ = f.SetCellStyle("Sheet1", "J"+strconv.Itoa(pos), "L"+strconv.Itoa(pos), topDotRightContinuousBorderUser)
			_ = f.SetCellStyle("Sheet1", "M"+strconv.Itoa(pos), "M"+strconv.Itoa(pos), topDotRightContinuousBorderSupervisor)
			_ = f.SetCellStyle("Sheet1", "N"+strconv.Itoa(pos), "N"+strconv.Itoa(pos), topDotLeftContinuousBorder)
		}

		_ = f.SetCellStyle("Sheet1", "A"+strconv.Itoa(pos), "A"+strconv.Itoa(pos), rightBorderBold)
		_ = f.SetCellStyle("Sheet1", "O"+strconv.Itoa(pos), "O"+strconv.Itoa(pos), leftBorderBold)

		if len(evalObj.Content.Challenges) == 0 || len(evalObj.Content.Challenges) == 1 || i == len(evalObj.Content.Challenges)-1 {
			_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(pos+1), "N"+strconv.Itoa(pos+1), topBorderBold)
		}

		_ = f.SetRowHeight("Sheet1", pos, 100)
	}

	idx += 3
	fifthCategories := map[string]string{
		"B" + strconv.Itoa(idx):   categories["CT42"],
		"B" + strconv.Itoa(idx+1): categories["CT43"],
		"J" + strconv.Itoa(idx+1): categories["CT44"],
	}
	for k, v := range fifthCategories {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	idx += 2
	_ = f.MergeCell("Sheet1", "B"+strconv.Itoa(idx), "H"+strconv.Itoa(idx+20))
	_ = f.MergeCell("Sheet1", "J"+strconv.Itoa(idx), "N"+strconv.Itoa(idx+20))

	_ = f.SetCellStyle("Sheet1", "B"+strconv.Itoa(idx), "H"+strconv.Itoa(idx+20), allThinBorderUser)
	_ = f.SetCellStyle("Sheet1", "J"+strconv.Itoa(idx), "N"+strconv.Itoa(idx+20), allThinBorderSupervisor)

	commentValue := map[string]string{
		"B" + strconv.Itoa(idx): evalObj.Content.Comment.SelfCmt,
		"J" + strconv.Itoa(idx): evalObj.Content.Comment.SuperiorCmt,
	}
	for k, v := range commentValue {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	buf, _ := f.WriteToBuffer()
	return buf.Bytes(), categories["CT1"] + "_Q" + strconv.Itoa(evalObj.Quarter) + strconv.Itoa(evalObj.Year) + "_" + evalObj.FullName
}

func (ctr *TargetEvalController) ExportEvaluationList(c echo.Context) error {
	userIdsStr := c.FormValue("user_ids")
	if userIdsStr == "" {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	var userIds []int
	err := json.Unmarshal([]byte(userIdsStr), &userIds)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	quarter, _ := strconv.Atoi(c.FormValue("quarter"))
	year, _ := strconv.Atoi(c.FormValue("year"))
	branch, _ := strconv.Atoi(c.FormValue("branch"))
	rank, _ := strconv.Atoi(c.FormValue("rank"))
	currentPage, _ := strconv.Atoi(c.FormValue("current_page"))

	exportExcelParams := &param.SearchEvaluationListParams{
		UserIds:     userIds,
		Name:        c.FormValue("name"),
		Quarter:     quarter,
		Year:        year,
		Branch:      branch,
		Rank:        rank,
		CurrentPage: currentPage,
		RowPerPage:  20,
	}

	userProfile := c.Get("user_profile").(m.User)
	evaluations, _, err := ctr.TargetEvaluationRepo.GetEvaluationList(userProfile.OrganizationID, exportExcelParams)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get Evaluation List Failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	f := excelize.NewFile()
	_ = f.SetColWidth("Sheet1", "A", "G", 25)

	titleStyle, _ := f.NewStyle(`{
		"font":{"bold":true, "size":16},
		"alignment":{"horizontal":"center", "vertical":"center"}
	}`)
	contentStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center", "vertical":"center"}}`)
	nameStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"left", "vertical":"center"}}`)

	_ = f.SetCellStyle("Sheet1", "A1", "G1", titleStyle)
	_ = f.SetColStyle("Sheet1", "A", contentStyle)
	_ = f.SetColStyle("Sheet1", "B", nameStyle)
	_ = f.SetColStyle("Sheet1", "C:G", contentStyle)

	var categories map[string]string
	if userProfile.LanguageId == cf.EnLanguageId {
		categories = cf.EvaluationListCategoriesEN
	} else if userProfile.LanguageId == cf.VnLanguageId {
		categories = cf.EvaluationListCategoriesVN
	} else {
		categories = cf.EvaluationListCategoriesJP
	}

	titleCategories := map[string]string{
		"A1": categories["employeeId"],
		"B1": categories["name"],
		"C1": categories["quarter"],
		"D1": categories["year"],
		"E1": categories["branch"],
		"F1": categories["point"],
		"G1": categories["rank"],
	}

	for k, v := range titleCategories {
		_ = f.SetCellValue("Sheet1", k, v)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	if len(branchRecords) > 0 {
		for _, record := range branchRecords {
			branches[record.Id] = record.Name
		}
	}

	for i, evaluation := range evaluations {
		pos := i + 2
		value := map[string]interface{}{
			"A" + strconv.Itoa(pos): evaluation.EmployeeId,
			"B" + strconv.Itoa(pos): evaluation.Name,
			"C" + strconv.Itoa(pos): evaluation.Quarter,
			"D" + strconv.Itoa(pos): evaluation.Year,
			"E" + strconv.Itoa(pos): branches[evaluation.Branch],
			"F" + strconv.Itoa(pos): evaluation.Point,
			"G" + strconv.Itoa(pos): cf.EvaluationRankList[evaluation.Rank],
		}

		for k, v := range value {
			_ = f.SetCellValue("Sheet1", k, v)
		}
	}

	buf, _ := f.WriteToBuffer()
	return c.Blob(http.StatusOK, "application/octet-stream", buf.Bytes())
}
