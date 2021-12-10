package projectboard

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgKanbanListRepository struct {
	cm.AppRepository
}

func NewPgKanbanListRepository(logger echo.Logger) (repo *PgKanbanListRepository) {
	repo = &PgKanbanListRepository{}
	repo.Init(logger)
	return
}

func (repo *PgKanbanListRepository) InsertKanbanList(params *param.CreateKanbanListParam) error {
	taskColumn := m.KanbanList{
		Name:            params.Name,
		KanbanBoardId:   params.KanbanBoardId,
		PositionInBoard: params.PositionInBoard,
	}

	err := repo.DB.Insert(&taskColumn)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanListRepository) UpdateKanbanList(params *param.EditKanbanListParam) error {
	var err error
	if params.Name != "" {
		_, err = repo.DB.Model(&m.KanbanList{Name: params.Name}).
			Column("name", "updated_at").
			Where("id = ?", params.Id).
			Where("kanban_board_id = ?", params.KanbanBoardId).
			Update()
	} else if len(params.SortPositionsInBoard) > 0 {
		err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
			var errTx error
			for _, pr := range params.SortPositionsInBoard {
				_, errTx := tx.Model(&m.KanbanList{PositionInBoard: pr.PositionInBoard}).
					Column("position_in_board", "updated_at").
					Where("id = ?", pr.Id).
					Update()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}
			}

			return errTx
		})
	}

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanListRepository) CountKanbanList(id int) (int, error) {
	count, err := repo.DB.Model(&m.KanbanList{}).
		Where("id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgKanbanListRepository) SelectKanbanLists(KanbanBoardId int) ([]m.KanbanList, error) {
	var KanbanLists []m.KanbanList
	err := repo.DB.Model(&KanbanLists).
		Column("id", "name").
		Where("kanban_board_id = ?", KanbanBoardId).
		Order("position_in_board ASC").
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return KanbanLists, err
}

func (repo *PgKanbanListRepository) DeleteKanbanList(params *param.RemoveKanbanListParam) error {
	_, err := repo.DB.Model(&m.KanbanList{}).
		Where("id = ?", params.Id).
		Where("kanban_board_id = ?", params.KanbanBoardId).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanListRepository) SelectProjectIdByKanbanListId(id int) (int, error) {
	var projectId int
	err := repo.DB.Model(&m.KanbanList{}).
		ColumnExpr("kb.project_id").
		Join("JOIN kanban_boards AS kb ON kb.id = kl.kanban_board_id").
		Where("kl.id = ?", id).
		Select(&projectId)

	if err != nil {
		repo.Logger.Error(err)
	}

	return projectId, err
}

func (repo *PgKanbanListRepository) SelectKanbanListNameByListId(id int) (string, error) {
	var name string
	err := repo.DB.Model(&m.KanbanList{}).
		ColumnExpr("name").
		Where("id = ?", id).
		Select(&name)

	if err != nil {
		repo.Logger.Error(err)
	}

	return name, err
}

