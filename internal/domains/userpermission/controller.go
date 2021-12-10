package userpermission

import (
	"net/http"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type Controller struct {
	cm.BaseController

	UserPermissionRepo rp.UserPermissionRepository
	UserRepo   rp.UserRepository
	OrgRepo    rp.OrgRepository
}

func NewUserPermissionController(logger echo.Logger, userPermissionRepo rp.UserPermissionRepository, userRepo rp.UserRepository, orgRepo rp.OrgRepository) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, userPermissionRepo, userRepo, orgRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) EditPermission(c echo.Context) error {
	editBranchParam := new(param.EditUserPermissionParam)
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

	err = ctr.UserPermissionRepo.UpdateUserPermission(userProfile.OrganizationID, editBranchParam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit user permission successfully.",
	})
}

func (ctr *Controller) GetPermissions(c echo.Context) error {
	selectUserPermissionParam := new(param.SelectUserPermissionParam)
	if err := c.Bind(selectUserPermissionParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.UserPermissionRepo.SelectPermissions(userProfile.OrganizationID, selectUserPermissionParam.UserId)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.SuccessResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	dataResponse := make(map[int]interface{})

	for i:= 1; i <= cf.TOTALMODULE; i++ {
		var module (map[string]interface{})
		var modules []map[string]interface{}

		for _, record := range records {
			if record.ModuleId == i {
				module = map[string]interface{}{
					"function_id": record.FunctionId,
					"status": cf.PermissionStatus[record.Status],
				}

				modules = append(modules, module)
				dataResponse[record.ModuleId] = modules
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get permissions successfully.",
		Data:    dataResponse,
	})
}

func (ctr *Controller) GetUserPermissions(c echo.Context) error {
	userPermissionParams := new(param.UserPermissionParams)
	if err := c.Bind(userPermissionParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}
	userProfile := c.Get("user_profile").(m.User)
	records, totalRow, err := ctr.UserPermissionRepo.SelectUserPermission(userProfile.OrganizationID, userPermissionParams)
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

	pagination := map[string]interface{}{
		"current_page": userPermissionParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": userPermissionParams.RowPerPage,
	}

	var responses []map[string]interface{}
	for _, record := range records {
		item := map[string]interface{}{
			"id":       	record.Id,
			"email":    	record.Email,
			"roleId": 		record.RoleID,
			"avatar":		record.Avatar,
			"first_name": 	record.FirstName,
			"last_name": 	record.LastName,
			"has_custom": 	record.HasCustom,
		}

		responses = append(responses, item)
	}

	dataResponse := map[string]interface{}{
		"pagination":       pagination,
		"user_permissions": responses,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get user permissions successfully.",
		Data:    dataResponse,
	})
}