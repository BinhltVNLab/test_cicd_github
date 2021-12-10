package kanbantask

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
	"strconv"
)

type PgKanbanTaskRepository struct {
	cm.AppRepository
}

func NewPgKanbanTaskRepository(logger echo.Logger) (repo *PgKanbanTaskRepository) {
	repo = &PgKanbanTaskRepository{}
	repo.Init(logger)
	return
}

func (repo *PgKanbanTaskRepository) InsertKanbanTask(currentUserId int, params *param.CreateKanbanTaskParam) error {
	task := m.KanbanTask{
		KanbanListId:   params.KanbanListId,
		Title:          params.Title,
		Status:         cf.UNDONE,
		PositionInList: params.PositionInList,
		CreatedBy:      currentUserId,
		UpdatedBy:      currentUserId,
	}

	err := repo.DB.Insert(&task)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanTaskRepository) UpdateKanbanTask(
	currentUserId int,
	organizationId int,
	params *param.EditKanbanTaskParam,
	notificationRepo rp.NotificationRepository,
	kanbanBoardRepo rp.KanbanBoardRepository,
	kanbanListRepo rp.KanbanListRepository,
	usersId []int,
) (string, string, []int, error) {
	kanbanTask := m.KanbanTask{UpdatedBy: currentUserId}
	var err error
	var content, link string
	var assignees []int
	if params.NewKanbanListId != 0 {
		if params.KanbanListId == params.NewKanbanListId {
			err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
				var errTx error
				for _, pr := range params.SortPositionList {
					_, errTx := tx.Model(&m.KanbanTask{PositionInList: pr.PositionInList}).
						Column("position_in_list", "updated_at").
						Where("id = ?", pr.Id).
						Update()

					if errTx != nil {
						repo.Logger.Error(errTx)
						return errTx
					}
				}

				return errTx
			})
		} else {
			err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
				var errTx error
				kanbanTask.KanbanListId = params.NewKanbanListId
				_, errTx = tx.Model(&kanbanTask).
					Column("kanban_list_id").
					Where("id = ?", params.Id).
					Update()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				kanbanListName, errTx := kanbanListRepo.SelectKanbanListNameByListId(params.KanbanListId)
				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				newKanbanListName, errTx := kanbanListRepo.SelectKanbanListNameByListId(params.NewKanbanListId)
				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				project, errTx := kanbanBoardRepo.SelectProjectIdAndNameByBoardId(params.KanbanBoardId)
				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				notificationParam := new(param.InsertNotificationParam)
				notificationParam.Content = "has move one task from " + kanbanListName +
					" to " + newKanbanListName + " at " + project.Name
				notificationParam.RedirectUrl = "/workflow/project-board?project_id=" +
					strconv.Itoa(project.Id) + "&id=" + strconv.Itoa(params.KanbanBoardId) + "&task_id=" + strconv.Itoa(params.Id)

				for _, userId := range usersId {
					if userId == currentUserId {
						continue
					}
					notificationParam.Receiver = userId

					errTx = notificationRepo.InsertNotificationWithTx(tx, organizationId, currentUserId, notificationParam)
					if errTx != nil {
						repo.Logger.Error(errTx)
						return errTx
					}
				}

				if len(params.SortPositionList) > 0 {
					for _, pr := range params.SortPositionList {
						_, errTx := tx.Model(&m.KanbanTask{PositionInList: pr.PositionInList}).
							Column("position_in_list", "updated_at").
							Where("id = ?", pr.Id).
							Update()

						if errTx != nil {
							repo.Logger.Error(errTx)
							return errTx
						}
					}
				}

				for _, pr := range params.SortNewPositionList {
					_, errTx := tx.Model(&m.KanbanTask{PositionInList: pr.PositionInList}).
						Column("position_in_list", "updated_at").
						Where("id = ?", pr.Id).
						Update()

					if errTx != nil {
						repo.Logger.Error(errTx)
						return errTx
					}
				}

				content = notificationParam.Content
				link = notificationParam.RedirectUrl

				return errTx
			})
		}
	} else {
		kanbanTask.Title = params.Title
		kanbanTask.Description = params.Description
		if params.DueDate != "" {
			kanbanTask.DueDate = calendar.ParseTime(cf.FormatDateNoSec, params.DueDate)
		}
		kanbanTask.Status = params.Status
		kanbanTask.Checklists = params.Checklists

		err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
			var errTx error
			assigneesTask, errTx := repo.SelectAssigneeByTaskId(params.Id)
			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}

			if !utils.IsEqualSlice(params.Assignees, assigneesTask) {
				kanbanTask.Assignees = params.Assignees
			}

			_, errTx = tx.Model(&kanbanTask).Where("id = ?", params.Id).UpdateNotZero()
			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}

			if len(params.Assignees) > 0 {
				project, errTx := kanbanBoardRepo.SelectProjectIdAndNameByBoardId(params.KanbanBoardId)
				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}

				assignees = utils.DiffSlice(assigneesTask, params.Assignees)
				if len(assignees) > 0 {
					content = "has assign a task for you at " + project.Name
					link = "/workflow/project-board?project_id=" + strconv.Itoa(project.Id) + "&id=" +
						strconv.Itoa(params.KanbanBoardId) + "&task_id=" + strconv.Itoa(params.Id)

					for _, assignee := range assignees {
						notificationParam := new(param.InsertNotificationParam)
						notificationParam.Receiver = assignee
						notificationParam.Content = content
						notificationParam.RedirectUrl = link

						errTx = notificationRepo.InsertNotificationWithTx(tx, organizationId, currentUserId, notificationParam)
						if errTx != nil {
							repo.Logger.Error(errTx)
							return errTx
						}
					}
				}
			}

			return errTx
		})
	}

	return content, link, assignees, err
}

func (repo *PgKanbanTaskRepository) CountKanbanTask(id int) (int, error) {
	count, err := repo.DB.Model(&m.KanbanTask{}).
		Where("id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgKanbanTaskRepository) SelectKanbanTasks(kanbanListId int) ([]param.KanbanTaskResponse, error) {
	var records []param.KanbanTaskResponse
	q := "SELECT n1.id AS kanban_task_id, n1.title, n1.assignees, n1.due_date, n1.status, n1.position_in_list, total_check_list, total_check_list_done " +
		"FROM (" +
		"SELECT id, title, assignees, due_date, status, kanban_list_id, position_in_list, deleted_at, jsonb_array_length(checklists) AS total_check_list " +
		"FROM kanban_tasks " +
		"GROUP BY id, title, assignees, kanban_list_id, position_in_list) n1 " +
		"LEFT JOIN (" +
		"SELECT id, COUNT(checklists) AS total_check_list_done " +
		"FROM kanban_tasks, json_array_elements(checklists::json) obj " +
		"WHERE obj ->> 'status' = 'true' " +
		"GROUP BY id) n2 ON n2.id = n1.id " +
		"WHERE n1.kanban_list_id = " + strconv.Itoa(kanbanListId) + " " +
		"AND n1.deleted_at IS NULL " +
		"ORDER BY n1.position_in_list ASC"

	_, err := repo.DB.Query(&records, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgKanbanTaskRepository) SelectKanbanTaskById(kanbanTaskId int) (m.KanbanTask, error) {
	var kanbanTask m.KanbanTask
	err := repo.DB.Model(&kanbanTask).
		Column("id", "title", "description", "assignees", "due_date", "status", "checklists").
		Where("id = ?", kanbanTaskId).
		Select(&kanbanTask)

	if err != nil {
		repo.Logger.Error(err)
	}

	return kanbanTask, err
}

func (repo *PgKanbanTaskRepository) DeleteKanbanTask(currentUserId int, id int) error {
	_, err := repo.DB.Model(&m.KanbanTask{UpdatedBy: currentUserId}).
		Column("updated_by", "updated_at").
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgKanbanTaskRepository) SelectAssigneeByTaskId(id int) ([]int, error) {
	var assignees []int
	err := repo.DB.Model(&m.KanbanTask{}).
		Column("assignees").
		Where("id = ?", id).
		Select(pg.Array(&assignees))

	if err != nil {
		repo.Logger.Error(err)
	}

	return assignees, err
}
