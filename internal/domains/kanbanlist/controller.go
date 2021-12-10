package projectboard

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"net/http"
)

type Controller struct {
	cm.BaseController

	KanbanListRepo rp.KanbanListRepository
	KanbanBoardRepo      rp.KanbanBoardRepository
	UserProjectRepo       rp.UserProjectRepository
}

func NewKanbanListController(
	logger echo.Logger,
	kanbanListRepo rp.KanbanListRepository,
	kanbanBoardRepo rp.KanbanBoardRepository,
	userProjectRepo rp.UserProjectRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, kanbanListRepo, kanbanBoardRepo, userProjectRepo}
	ctr.Init(logger)

	return
}

func (ctr *Controller) CreateKanbanList(c echo.Context) error {
	params := new(param.CreateKanbanListParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	count, err := ctr.KanbanBoardRepo.CountKanbanBoard(params.KanbanBoardId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Kanban board does not exist",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID {
		userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, params.ProjectId)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		var usersId []int
		if len(userRecords) > 0 {
			for _, user := range userRecords {
				usersId = append(usersId, user.UserId)
			}
		}

		if !utils.FindIntInSlice(usersId, userProfile.UserProfile.UserID) {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not a member of the project",
			})
		}
	}

	if err := ctr.KanbanListRepo.InsertKanbanList(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create kanban list successful",
	})
}

func (ctr *Controller) EditKanbanList(c echo.Context) error {
	params := new(param.EditKanbanListParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if params.Name != "" && params.Id == 0 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Id is required",
		})
	}

	count, err := ctr.KanbanBoardRepo.CountKanbanBoard(params.KanbanBoardId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Kanban board does not exist",
		})
	}

	if params.Id != 0 {
		count, err = ctr.KanbanListRepo.CountKanbanList(params.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		if count == 0 {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Kanban list does not exist",
			})
		}
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID {
		userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, params.ProjectId)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		var usersId []int
		if len(userRecords) > 0 {
			for _, user := range userRecords {
				usersId = append(usersId, user.UserId)
			}
		}

		if !utils.FindIntInSlice(usersId, userProfile.UserProfile.UserID) {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not a member of the project",
			})
		}
	}

	if err := ctr.KanbanListRepo.UpdateKanbanList(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit kanban list successful",
	})
}

func (ctr *Controller) RemoveKanbanList(c echo.Context) error {
	params := new(param.RemoveKanbanListParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	count, err := ctr.KanbanBoardRepo.CountKanbanBoard(params.KanbanBoardId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Kanban board does not exist",
		})
	}

	count, err = ctr.KanbanListRepo.CountKanbanList(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Kanban list does not exist",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	if userProfile.RoleID != cf.GeneralManagerRoleID {
		userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, params.ProjectId)
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		var usersId []int
		if len(userRecords) > 0 {
			for _, user := range userRecords {
				usersId = append(usersId, user.UserId)
			}
		}

		if !utils.FindIntInSlice(usersId, userProfile.UserProfile.UserID) {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not a member of the project",
			})
		}
	}

	if err := ctr.KanbanListRepo.DeleteKanbanList(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove kanban successful",
	})
}
