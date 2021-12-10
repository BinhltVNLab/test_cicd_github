package usertechnology

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

	UserTechnologyRepo rp.UserTechnologyRepository
	TechnologyRepo     rp.TechnologyRepository
	UserRepo           rp.UserRepository
}

// NewUserTechnologyController : Init UserTechnology Controller
func NewUserTechnologyController(
	logger echo.Logger,
	userTechnologyRepo rp.UserTechnologyRepository,
	technologyRepo rp.TechnologyRepository,
	userRepo rp.UserRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, userTechnologyRepo, technologyRepo, userRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateUserTechnologies(c echo.Context) error {
	createUserTechnologyParams := new([]param.CreateUserTechnologyParam)
	if err := c.Bind(createUserTechnologyParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	for _, createParam := range *createUserTechnologyParams {
		_, err := valid.ValidateStruct(createParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
			})
		}

		if userProfile.RoleID != cf.GeneralManagerRoleID && userProfile.RoleID != cf.ManagerRoleID && userProfile.UserProfile.UserID != createParam.UserId {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You not have permission to set technologies.",
			})
		}

		checkExistTechnology, err := ctr.TechnologyRepo.CheckExistTechnology(userProfile.OrganizationID, createParam.TechnologyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if checkExistTechnology == 0 {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Technology is not exist.",
			})
		}

		count, err := ctr.UserTechnologyRepo.CheckExistUserTechnology(userProfile.OrganizationID, createParam.UserId, createParam.TechnologyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if count > 0 {
			return c.JSON(http.StatusNotAcceptable, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "This technology already exist",
			})
		}

		err = ctr.UserTechnologyRepo.InsertUserTechnology(createParam.UserId, createParam.TechnologyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "User add technologies successfully.",
	})
}

func (ctr *Controller) GetTechnologiesOfUser(c echo.Context) error {
	getTechnologiesOfUserParam := new(param.GetTechnologiesOfUserParam)
	if err := c.Bind(getTechnologiesOfUserParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(getTechnologiesOfUserParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.UserTechnologyRepo.SelectTechnologiesByUserId(userProfile.OrganizationID, getTechnologiesOfUserParam.UserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}
	if len(records) == 0 {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Empty records",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get technologies of this user successfully.",
		Data:    records,
	})
}

func (ctr *Controller) RemoveTechnologiesOfUser(c echo.Context) error {
	removeTechnologyOfUserParams := new([]param.RemoveTechnologyOfUserParam)
	if err := c.Bind(removeTechnologyOfUserParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	for _, removeParam := range *removeTechnologyOfUserParams {
		_, err := valid.ValidateStruct(removeParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid field value",
			})
		}

		count, err := ctr.UserTechnologyRepo.CheckExistUserTechnologyById(userProfile.OrganizationID, removeParam.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if count == 0 {
			return c.JSON(http.StatusNotAcceptable, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Technology is not exists",
			})
		}

		if userProfile.RoleID != cf.GeneralManagerRoleID && userProfile.RoleID != cf.ManagerRoleID && userProfile.UserProfile.UserID != removeParam.UserId {
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You not have permission to remove technology.",
			})
		}

		err = ctr.UserTechnologyRepo.DeleteUserTechnologyById(userProfile.OrganizationID, removeParam.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove technology successfully.",
	})
}

func (ctr *Controller) TechnologyStatisticDetail(c echo.Context) error {
	technologyStatisticDetailParams := new(param.TechnologyStatisticDetailParams)
	if err := c.Bind(technologyStatisticDetailParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(technologyStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, totalRow, err := ctr.UserTechnologyRepo.SelectUsersByTechnologyName(userProfile.OrganizationID, technologyStatisticDetailParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	pagination := map[string]interface{}{
		"current_page": technologyStatisticDetailParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": technologyStatisticDetailParams.RowPerPage,
	}

	var responses []map[string]interface{}
	for _, record := range records {
		res := map[string]interface{}{
			"user_id":             record.UserId,
			"full_name":           record.FullName,
			"job_title":           record.JobTitle,
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
		Message: "Get technology statistic detail successful",
		Data:    dataResponse,
	})
}
