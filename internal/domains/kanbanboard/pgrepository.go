package kanbanboard

import (
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgKanbanBoardRepository struct {
	cm.AppRepository
}

func NewPgKanbanBoardRepository(logger echo.Logger) (repo *PgKanbanBoardRepository) {
	repo = &PgKanbanBoardRepository{}
	repo.Init(logger)
	return
}

func (repo *PgKanbanBoardRepository) InsertKanbanBoard(params *param.CreateKanbanBoardParam) error {
	KanbanBoard := m.KanbanBoard{
		Name:      params.Name,
		ProjectId: params.ProjectId,
	}

	err := repo.DB.Insert(&KanbanBoard)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanBoardRepository) UpdateKanbanBoard(params *param.EditKanbanBoardParam) error {
	_, err := repo.DB.Model(&m.KanbanBoard{Name: params.Name}).
		Column("name", "updated_at").
		Where("id = ?", params.Id).
		Where("project_id = ?", params.ProjectId).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanBoardRepository) CountKanbanBoard(id int) (int, error) {
	count, err := repo.DB.Model(&m.KanbanBoard{}).
		Where("id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgKanbanBoardRepository) SelectKanbanBoards(projectId int) ([]m.KanbanBoard, error) {
	var boards []m.KanbanBoard
	err := repo.DB.Model(&boards).
		Column("id", "name").
		Where("project_id = ?", projectId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return boards, err
}

func (repo *PgKanbanBoardRepository) DeleteKanbanBoard(params *param.RemoveKanbanBoardParam) error {
	_, err := repo.DB.Model(&m.KanbanBoard{}).
		Where("id = ?", params.Id).
		Where("project_id = ?", params.ProjectId).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanBoardRepository) SelectKanbanBoardById(columns []string, id int) (m.KanbanBoard, error) {
	var KanbanBoard m.KanbanBoard
	err := repo.DB.Model(&KanbanBoard).
		Column(columns...).
		Where("id = ?", id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return KanbanBoard, err
}

func (repo *PgKanbanBoardRepository) SelectProjectIdAndNameByBoardId(kanbanBoardId int) (param.ProjectIdAndNameRecord, error) {
	var project param.ProjectIdAndNameRecord
	err := repo.DB.Model(&m.KanbanBoard{}).
		Column("p.id", "p.name").
		Join("JOIN projects AS p ON p.id = kb.project_id").
		Where("kb.id = ?", kanbanBoardId).
		Select(&project)

	if err != nil {
		repo.Logger.Error(err)
	}

	return project, err
}
