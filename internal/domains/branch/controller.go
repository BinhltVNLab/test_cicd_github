package branch

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"net/http"
)

type Controller struct {
	cm.BaseController

	BranchRepo rp.BranchRepository
	UserRepo   rp.UserRepository
	OrgRepo    rp.OrgRepository
}

func NewBranchController(logger echo.Logger, branchRepo rp.BranchRepository, userRepo rp.UserRepository, orgRepo rp.OrgRepository) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, branchRepo, userRepo, orgRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateBranch(c echo.Context) error {
	createBranchParam := new(param.CreateBranchParam)
	if err := c.Bind(createBranchParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createBranchParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.BranchRepo.CheckExistBranchByName(userProfile.OrganizationID, createBranchParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The branch already exist.",
		})
	}

	err = ctr.BranchRepo.InsertBranch(
		userProfile.OrganizationID,
		createBranchParam,
		userProfile.Organization.SettingStep,
		ctr.OrgRepo,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create branch successfully.",
	})
}

func (ctr *Controller) EditBranch(c echo.Context) error {
	editBranchParam := new(param.EditBranchParam)
	if err := c.Bind(editBranchParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(editBranchParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.BranchRepo.CheckExistBranch(userProfile.OrganizationID, editBranchParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Empty records.",
		})
	}

	count, err = ctr.BranchRepo.CheckExistBranchByName(userProfile.OrganizationID, editBranchParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The branch already exist.",
		})
	}

	err = ctr.BranchRepo.UpdateBranch(userProfile.OrganizationID, editBranchParam.Id, editBranchParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit branch successfully.",
	})
}

func (ctr *Controller) GetBranches(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
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

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get branches successfully.",
		Data:    records,
	})
}

func (ctr *Controller) RemoveBranch(c echo.Context) error {
	removeBranchParam := new(param.RemoveBranchParam)
	if err := c.Bind(removeBranchParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(removeBranchParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.BranchRepo.CheckExistBranch(userProfile.OrganizationID, removeBranchParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Empty records.",
		})
	}

	err = ctr.UserRepo.UpdateIntFieldToNull(userProfile.OrganizationID, "branch", removeBranchParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	err = ctr.BranchRepo.DeleteBranch(userProfile.OrganizationID, removeBranchParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Delete branch successfully.",
	})
}

func (ctr *Controller) BranchStatisticDetail(c echo.Context) error {
	branchStatisticDetailParams := new(param.BranchStatisticDetailParams)
	if err := c.Bind(branchStatisticDetailParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(branchStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, totalRow, err := ctr.BranchRepo.SelectUsersByBranchName(userProfile.OrganizationID, branchStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": branchStatisticDetailParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": branchStatisticDetailParams.RowPerPage,
	}

	var responses []map[string]interface{}
	for _, record := range records {
		res := map[string]interface{}{
			"user_id":             record.UserId,
			"full_name":           record.FullName,
			"job_title":           record.JobTitle,
			"birthday":            record.Birthday.Format(cf.FormatDateDisplay),
			"company_joined_date": record.CompanyJoinedDate.Format(cf.FormatDateDisplay),
		}
		responses = append(responses, res)
	}

	dataResponse := map[string]interface{}{
		"pagination":       pagination,
		"statistic_detail": responses,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get branch statistic detail successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) SortPrioritization(c echo.Context) error {
	params := new(param.SortPrioritizationParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	count, err := ctr.BranchRepo.CountBranches(params.OrganizationId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if len(params.Priorities) != count {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	keys := make(map[int]bool)
	for _, pr := range params.Priorities {
		if _, err := valid.ValidateStruct(pr); err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
			})
		}

		if _, value := keys[pr.Priority]; !value {
			keys[pr.Priority] = true
		} else {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
			})
		}
	}

	if err := ctr.BranchRepo.UpdatePriorityWithTx(params.Priorities); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update priority successful",
	})
}
