package jobtitle

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

	JobTitleRepo rp.JobTitleRepository
	UserRepo     rp.UserRepository
	OrgRepo      rp.OrgRepository
}

func NewJobTitleController(logger echo.Logger, jobTitleRepo rp.JobTitleRepository, userRepo rp.UserRepository, orgRepo rp.OrgRepository) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, jobTitleRepo, userRepo, orgRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateJobTitle(c echo.Context) error {
	createJobTitleParam := new(param.CreateJobTitleParam)
	if err := c.Bind(createJobTitleParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createJobTitleParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.JobTitleRepo.CheckExistJobTitleByName(userProfile.OrganizationID, createJobTitleParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The job title already exist.",
		})
	}

	err = ctr.JobTitleRepo.InsertJobTitle(
		userProfile.OrganizationID,
		createJobTitleParam,
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
		Message: "Create job title successfully.",
	})
}

func (ctr *Controller) EditJobTitle(c echo.Context) error {
	editJobTitleParam := new(param.EditJobTitleParam)
	if err := c.Bind(editJobTitleParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(editJobTitleParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.JobTitleRepo.CheckExistJobTitle(userProfile.OrganizationID, editJobTitleParam.Id)
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

	count, err = ctr.JobTitleRepo.CheckExistJobTitleByName(userProfile.OrganizationID, editJobTitleParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The job title already exist.",
		})
	}

	err = ctr.JobTitleRepo.UpdateJobTitle(userProfile.OrganizationID, editJobTitleParam.Id, editJobTitleParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit job title successfully.",
	})
}

func (ctr *Controller) GetJobTitles(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.JobTitleRepo.SelectJobTitles(userProfile.OrganizationID)
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
		Message: "Get job titles successfully.",
		Data:    records,
	})
}

func (ctr *Controller) RemoveJobTitle(c echo.Context) error {
	removeJobTitleParam := new(param.RemoveJobTitleParam)
	if err := c.Bind(removeJobTitleParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(removeJobTitleParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.JobTitleRepo.CheckExistJobTitle(userProfile.OrganizationID, removeJobTitleParam.Id)
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

	err = ctr.UserRepo.UpdateIntFieldToNull(userProfile.OrganizationID, "job_title", removeJobTitleParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	err = ctr.JobTitleRepo.DeleteJobTitle(userProfile.OrganizationID, removeJobTitleParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Delete job title successfully.",
	})
}

func (ctr *Controller) JobTitleStatisticDetail(c echo.Context) error {
	jobTitleStatisticDetailParams := new(param.JobTitleStatisticDetailParams)
	if err := c.Bind(jobTitleStatisticDetailParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(jobTitleStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, totalRow, err := ctr.JobTitleRepo.SelectUsersByJobTitleName(userProfile.OrganizationID, jobTitleStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": jobTitleStatisticDetailParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": jobTitleStatisticDetailParams.RowPerPage,
	}

	var responses []map[string]interface{}
	for _, record := range records {
		res := map[string]interface{}{
			"user_id":             record.UserId,
			"full_name":           record.FullName,
			"branch":              record.Branch,
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
		Message: "Get job title statistic detail successful",
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

	count, err := ctr.JobTitleRepo.CountJobTitles(params.OrganizationId)
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

	if err := ctr.JobTitleRepo.UpdatePriorityWithTx(params.Priorities); err != nil {
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
