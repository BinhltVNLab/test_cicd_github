package technology

import (
	valid "github.com/asaskevich/govalidator"
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

	TechnologyRepo     rp.TechnologyRepository
	UserTechnologyRepo rp.UserTechnologyRepository
	OrgRepo            rp.OrgRepository
}

func NewTechnologyController(
	logger echo.Logger,
	technologyRepo rp.TechnologyRepository,
	userTechnologyRepo rp.UserTechnologyRepository,
	orgRepo rp.OrgRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, technologyRepo, userTechnologyRepo, orgRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateTechnology(c echo.Context) error {
	createTechnologyParam := new(param.CreateTechnologyParam)
	if err := c.Bind(createTechnologyParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createTechnologyParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.TechnologyRepo.CheckExistTechnologyByName(userProfile.OrganizationID, createTechnologyParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The technology already exist.",
		})
	}

	err = ctr.TechnologyRepo.InsertTechnology(
		userProfile.OrganizationID,
		createTechnologyParam,
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
		Message: "Create technology successfully.",
	})
}

func (ctr *Controller) EditTechnology(c echo.Context) error {
	editTechnologyParam := new(param.EditTechnologyParam)
	if err := c.Bind(editTechnologyParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(editTechnologyParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.TechnologyRepo.CheckExistTechnology(userProfile.OrganizationID, editTechnologyParam.Id)
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
	count, err = ctr.TechnologyRepo.CheckExistTechnologyByName(userProfile.OrganizationID, editTechnologyParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The technology already exist.",
		})
	}

	err = ctr.TechnologyRepo.UpdateTechnology(userProfile.OrganizationID, editTechnologyParam.Id, editTechnologyParam.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit technology successfully.",
	})
}

func (ctr *Controller) GetTechnologies(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.TechnologyRepo.SelectTechnologies(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get technologies successfully.",
		Data:    records,
	})
}

func (ctr *Controller) RemoveTechnology(c echo.Context) error {
	removeTechnologyParam := new(param.RemoveTechnologyParam)
	if err := c.Bind(removeTechnologyParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(removeTechnologyParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.TechnologyRepo.CheckExistTechnology(userProfile.OrganizationID, removeTechnologyParam.Id)
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

	err = ctr.UserTechnologyRepo.DeleteUserTechnologyByTechnologyId(userProfile.OrganizationID, removeTechnologyParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	err = ctr.TechnologyRepo.DeleteTechnology(userProfile.OrganizationID, removeTechnologyParam.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Delete technology successfully.",
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

	count, err := ctr.TechnologyRepo.CountTechnologies(params.OrganizationId)
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

	if err := ctr.TechnologyRepo.UpdatePriorityWithTx(params.Priorities); err != nil {
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
