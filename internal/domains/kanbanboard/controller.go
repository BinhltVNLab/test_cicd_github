package kanbanboard

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

	KanbanBoardRepo rp.KanbanBoardRepository
	ProjectRepo     rp.ProjectRepository
	UserProjectRepo rp.UserProjectRepository
}

func NewKanbanBoardController(
	logger echo.Logger,
	kanbanBoardRepo rp.KanbanBoardRepository,
	projectRepo rp.ProjectRepository,
	userProjectRepo rp.UserProjectRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, kanbanBoardRepo, projectRepo, userProjectRepo}
	ctr.Init(logger)

	return
}

func (ctr *Controller) CreateKanbanBoard(c echo.Context) error {
	params := new(param.CreateKanbanBoardParam)
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

	count, err := ctr.ProjectRepo.CountProjectById(params.ProjectId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project does not exist",
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
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You are not a member of the project",
			})
		}
	}

	if err := ctr.KanbanBoardRepo.InsertKanbanBoard(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create kanban board successful",
	})
}

func (ctr *Controller) EditKanbanBoard(c echo.Context) error {
	params := new(param.EditKanbanBoardParam)
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

	count, err := ctr.ProjectRepo.CountProjectById(params.ProjectId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project does not exist",
		})
	}

	count, err = ctr.KanbanBoardRepo.CountKanbanBoard(params.Id)
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
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You are not a member of the project",
			})
		}
	}

	if err := ctr.KanbanBoardRepo.UpdateKanbanBoard(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit kanban board successful",
	})
}

func (ctr *Controller) GetKanbanBoards(c echo.Context) error {
	params := new(param.GetKanbanBoardsParam)
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

	count, err := ctr.ProjectRepo.CountProjectById(params.ProjectId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project does not exist",
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
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You are not a member of the project",
			})
		}
	}

	boards, err := ctr.KanbanBoardRepo.SelectKanbanBoards(params.ProjectId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var dataResponse []map[string]interface{}
	for _, board := range boards {
		data := map[string]interface{}{
			"id":   board.ID,
			"name": board.Name,
		}

		dataResponse = append(dataResponse, data)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get kanban boards successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) RemoveKanbanBoard(c echo.Context) error {
	params := new(param.RemoveKanbanBoardParam)
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

	count, err := ctr.ProjectRepo.CountProjectById(params.ProjectId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project does not exist",
		})
	}

	count, err = ctr.KanbanBoardRepo.CountKanbanBoard(params.Id)
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
			return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "You are not a member of the project",
			})
		}
	}

	if err := ctr.KanbanBoardRepo.DeleteKanbanBoard(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove kanban board successful",
	})
}
