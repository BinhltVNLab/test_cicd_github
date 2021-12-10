package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type KanbanListRepository interface {
	InsertKanbanList(params *param.CreateKanbanListParam) error
	CountKanbanList(id int) (int, error)
	UpdateKanbanList(params *param.EditKanbanListParam) error
	SelectKanbanLists(boardId int) ([]m.KanbanList, error)
	DeleteKanbanList(params *param.RemoveKanbanListParam) error
	SelectProjectIdByKanbanListId(id int) (int, error)
	SelectKanbanListNameByListId(id int) (string, error)
}
