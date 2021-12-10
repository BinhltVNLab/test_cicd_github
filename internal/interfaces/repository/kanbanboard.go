package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type KanbanBoardRepository interface {
	InsertKanbanBoard(params *param.CreateKanbanBoardParam) error
	CountKanbanBoard(id int) (int, error)
	UpdateKanbanBoard(params *param.EditKanbanBoardParam) error
	SelectKanbanBoards(projectId int) ([]m.KanbanBoard, error)
	DeleteKanbanBoard(params *param.RemoveKanbanBoardParam) error
	SelectKanbanBoardById(columns []string, id int) (m.KanbanBoard, error)
	SelectProjectIdAndNameByBoardId(kanbanBoardId int) (param.ProjectIdAndNameRecord, error)
}
