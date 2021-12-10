package kanbantask

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	afb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/appfirebase"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"net/http"
	"time"
)

type Controller struct {
	cm.BaseController
	afb.FirebaseCloudMessage

	Cloud            gc.StorageUtility
	KanbanTaskRepo   rp.KanbanTaskRepository
	KanbanListRepo   rp.KanbanListRepository
	KanbanBoardRepo  rp.KanbanBoardRepository
	UserProjectRepo  rp.UserProjectRepository
	ProjectRepo      rp.ProjectRepository
	NotificationRepo rp.NotificationRepository
	FcmTokenRepo     rp.FcmTokenRepository
	UserRepo         rp.UserRepository
}

func NewKanbanTaskController(
	logger echo.Logger,
	cloud gc.StorageUtility,
	kanbanTaskRepo rp.KanbanTaskRepository,
	kanbanListRepo rp.KanbanListRepository,
	kanbanBoardRepo rp.KanbanBoardRepository,
	userProjectRepo rp.UserProjectRepository,
	projectRepo rp.ProjectRepository,
	notificationRepo rp.NotificationRepository,
	fcmTokenRepo rp.FcmTokenRepository,
	userRepo rp.UserRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, afb.FirebaseCloudMessage{}, cloud, kanbanTaskRepo,
		kanbanListRepo, kanbanBoardRepo, userProjectRepo,
		projectRepo, notificationRepo, fcmTokenRepo, userRepo,
	}
	ctr.Init(logger)
	ctr.InitFcm()

	return
}

func (ctr *Controller) CreateKanbanTask(c echo.Context) error {
	params := new(param.CreateKanbanTaskParam)
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

	if err := ctr.KanbanTaskRepo.InsertKanbanTask(userProfile.UserProfile.UserID, params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create kanban task successful",
	})
}

func (ctr *Controller) EditKanbanTask(c echo.Context) error {
	params := new(param.EditKanbanTaskParam)
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

	if params.NewKanbanListId == 0 && params.Id == 0 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Id is required",
		})
	}

	if params.NewKanbanListId != 0 && (params.KanbanListId != params.NewKanbanListId && len(params.SortNewPositionList) == 0) {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if params.Id != 0 {
		count, err := ctr.KanbanTaskRepo.CountKanbanTask(params.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
			})
		}

		if count == 0 {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Kanban task does not exist",
			})
		}
	}

	userProfile := c.Get("user_profile").(m.User)
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

	if userProfile.RoleID != cf.GeneralManagerRoleID {
		if !utils.FindIntInSlice(usersId, userProfile.UserProfile.UserID) {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User is not a member of the project",
			})
		}

		if len(params.Assignees) > 0 {
			for _, assignee := range params.Assignees {
				if !utils.FindIntInSlice(usersId, assignee) {
					return c.JSON(http.StatusNotFound, cf.JsonResponse{
						Status:  cf.FailResponseCode,
						Message: "User is not a member of the project",
					})
				}
			}
		}
	}

	content, link, assignees, err := ctr.KanbanTaskRepo.UpdateKanbanTask(
		userProfile.UserProfile.UserID,
		userProfile.OrganizationID,
		params,
		ctr.NotificationRepo,
		ctr.KanbanBoardRepo,
		ctr.KanbanListRepo,
		usersId,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if content != "" && link != "" {
		var registrationTokens []string
		if params.NewKanbanListId != 0 {
			registrationTokens, err = ctr.FcmTokenRepo.SelectMultiFcmTokens(usersId, userProfile.UserProfile.UserID)
		} else if len(assignees) > 0 {
			registrationTokens, err = ctr.FcmTokenRepo.SelectMultiFcmTokens(assignees, userProfile.UserProfile.UserID)
		}

		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System error",
			})
		}

		if len(registrationTokens) > 0 {
			content = userProfile.UserProfile.FirstName + " " + userProfile.UserProfile.LastName + " " + content
			for _, token := range registrationTokens {
				err := ctr.SendMessageToSpecificUser(token, "Micro Erp New Notification", content, link)
				if err != nil {
					if err.Error() == "http error status: 400; reason: request contains an invalid argument; "+
						"code: invalid-argument; details: The registration token is not a valid FCM registration token" {
						err = ctr.FcmTokenRepo.DeleteFcmToken(token)
						if err != nil {
							ctr.Logger.Error(err)
						}
					} else {
						ctr.Logger.Error(err)
					}
				}
			}
		}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Edit kanban task status successful",
	})
}

func (ctr *Controller) GetKanbanTasks(c echo.Context) error {
	params := new(param.GetKanbanTasksParam)
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

	userProfile := c.Get("user_profile").(m.User)
	userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, params.ProjectId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var usersId []int
	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			usersId = append(usersId, user.UserId)
			users[user.UserId] = user.FullName
		}
	}

	if userProfile.RoleID != cf.GeneralManagerRoleID && !utils.FindIntInSlice(usersId, userProfile.UserProfile.UserID) {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "User is not a member of the project",
		})
	}

	columns := []string{"name"}
	kanbanBoard, err := ctr.KanbanBoardRepo.SelectKanbanBoardById(columns, params.KanbanBoardId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	kanbanLists, err := ctr.KanbanListRepo.SelectKanbanLists(params.KanbanBoardId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var kanbanListsResponse []map[string]interface{}
	if len(kanbanLists) > 0 {
		location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
		for _, kanbanList := range kanbanLists {
			records, err := ctr.KanbanTaskRepo.SelectKanbanTasks(kanbanList.ID)
			if err != nil && err.Error() != pg.ErrNoRows.Error() {
				return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "System Error",
				})
			}

			var kanbanTasksResponse []map[string]interface{}
			if len(records) > 0 {
				for _, record := range records {
					kanbanTaskElm := map[string]interface{}{
						"kanban_task_id":        record.KanbanTaskId,
						"title":                 record.Title,
						"assignees":             record.Assignees,
						"position_in_list":      record.PositionInList,
						"status":                record.Status,
						"total_check_list":      record.TotalCheckList,
						"total_check_list_done": record.TotalCheckListDone,
						"avatars":               ctr.SelectAssigneeAndAvatars(record.Assignees),
					}

					if err != nil {
						ctr.Logger.Error(err)
						return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
							Status:  cf.FailResponseCode,
							Message: "System Error",
						})
					}

					if record.DueDate.IsZero() {
						kanbanTaskElm["due_date"] = ""
					} else {
						kanbanTaskElm["due_date"] = record.DueDate.In(location).Format(cf.FormatDateNoSec)
					}

					kanbanTasksResponse = append(kanbanTasksResponse, kanbanTaskElm)
				}
			}

			kanbanListElm := map[string]interface{}{
				"kanban_list_id":   kanbanList.ID,
				"kanban_list_name": kanbanList.Name,
				"kanban_tasks":     kanbanTasksResponse,
			}

			kanbanListsResponse = append(kanbanListsResponse, kanbanListElm)
		}
	}

	dataResponse := map[string]interface{}{
		"kanban_board_name": kanbanBoard.Name,
		"kanban_lists":      kanbanListsResponse,
		"status_list":       cf.TaskStatusMap,
		"users":             users,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get kanban tasks successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) GetKanbanTask(c echo.Context) error {
	params := new(param.GetKanbanTaskParam)
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

	record, err := ctr.KanbanTaskRepo.SelectKanbanTaskById(params.KanbanTaskId)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusNotFound, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Task does not exist",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, params.ProjectId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var usersId []int
	users := make(map[int]string)
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			usersId = append(usersId, user.UserId)
			users[user.UserId] = user.FullName
		}
	}

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	kanbanTask := map[string]interface{}{
		"id":          record.ID,
		"title":       record.Title,
		"description": record.Description,
		"assignees":   record.Assignees,
		"status":      record.Status,
		"checklists":  record.Checklists,
	}

	if record.DueDate.IsZero() {
		kanbanTask["due_date"] = ""
	} else {
		kanbanTask["due_date"] = record.DueDate.In(location).Format(cf.FormatDateNoSec)
	}

	if err != nil {
		ctr.Logger.Error(err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	dataResponse := map[string]interface{}{
		"kanban_task": kanbanTask,
		"users":       users,
		"avatars":     ctr.SelectAssigneeAndAvatars(usersId),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get task successful",
		Data:    dataResponse,
	})
}

func (ctr *Controller) RemoveKanbanTask(c echo.Context) error {
	params := new(param.RemoveKanbanTaskParam)
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

	count, err := ctr.KanbanTaskRepo.CountKanbanTask(params.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusNotFound, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Kanban task does not exist",
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

	if err := ctr.KanbanTaskRepo.DeleteKanbanTask(userProfile.UserProfile.UserID, params.Id); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove kanban task successful",
	})
}

func (ctr *Controller) SelectAssigneeAndAvatars(assignees []int) map[int][]byte {
	records := make(map[int][]byte)
	if len(assignees) == 0 {
		return records
	}

	userIdAndAvatars, err := ctr.UserRepo.SelectAvatarUsers(assignees)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		panic(err)
	}

	for _, user := range userIdAndAvatars {
		var base64Img []byte
		if user.Avatar != "" {
			base64Img, err = ctr.Cloud.GetFileByFileName(user.Avatar, cf.AvatarFolderGCS)
			if err != nil {
				ctr.Logger.Error(err)
			}
		}

		records[user.UserId] = base64Img
	}

	return records
}
