package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type KanbanTaskRepository interface {
	InsertKanbanTask(currentUserId int, params *param.CreateKanbanTaskParam) error
	UpdateKanbanTask(
		currentUserId int,
		organizationId int,
		params *param.EditKanbanTaskParam,
		notificationRepo NotificationRepository,
		kanbanBoardRepo KanbanBoardRepository,
		kanbanListRepo KanbanListRepository,
		usersId []int,
	) (string, string, []int, error)
	SelectKanbanTasks(kanbanListId int) ([]param.KanbanTaskResponse, error)
	SelectKanbanTaskById(kanbanTaskId int) (m.KanbanTask, error)
	CountKanbanTask(id int) (int, error)
	DeleteKanbanTask(currentUserId int, id int) error
}
