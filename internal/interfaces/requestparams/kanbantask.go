package requestparams

import (
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"time"
)

type CreateKanbanTaskParam struct {
	ProjectId      int    `json:"project_id" valid:"required"`
	KanbanListId   int    `json:"kanban_list_id" valid:"required"`
	Title          string `json:"title" valid:"required"`
	PositionInList int    `json:"position_in_list" valid:"required"`
}

type EditKanbanTaskParam struct {
	Id                  int                       `json:"id"`
	KanbanListId        int                       `json:"kanban_list_id" valid:"required"`
	KanbanBoardId       int                       `json:"kanban_board_id"`
	ProjectId           int                       `json:"project_id" valid:"required"`
	Title               string                    `json:"title"`
	Description         string                    `json:"description"`
	DueDate             string                    `json:"due_date"`
	Assignees           []int                     `json:"assignees"`
	Status              int                       `json:"status" valid:"range(1|2)"`
	Checklists          []m.Checklist             `json:"checklists"`
	NewKanbanListId     int                       `json:"new_kanban_list_id"`
	SortPositionList    []SortPositionInListParam `json:"sort_position_list"`
	SortNewPositionList []SortPositionInListParam `json:"sort_new_position_list"`
}

type GetKanbanTasksParam struct {
	KanbanBoardId int `json:"kanban_board_id" valid:"required"`
	ProjectId     int `json:"project_id" valid:"required"`
}

type GetKanbanTaskParam struct {
	KanbanTaskId int `json:"kanban_task_id" valid:"required"`
	ProjectId    int `json:"project_id" valid:"required"`
}

type SelectKanbanTaskRecords struct {
	Id             int       `json:"id"`
	Title          string    `json:"title"`
	Assignees      []int     `json:"assignees"`
	DueDate        time.Time `json:"due_date"`
	PositionInList int       `json:"position_in_list"`
}

type RemoveKanbanTaskParam struct {
	ProjectId    int `json:"project_id" valid:"required"`
	Id           int `json:"id" valid:"required"`
}

type KanbanTaskResponse struct {
	KanbanTaskId       int       `json:"kanban_task_id"`
	Title              string    `json:"title"`
	Assignees          []int     `json:"assignees" pg:",array"`
	DueDate            time.Time `json:"due_date"`
	PositionInList     int       `json:"position_in_list"`
	Status             int       `json:"status"`
	TotalCheckList     int       `json:"total_check_list"`
	TotalCheckListDone int       `json:"total_check_list_done"`
}

type SortPositionInListParam struct {
	Id             int `json:"id" valid:"required"`
	PositionInList int `json:"position_in_list" valid:"required"`
}

type UserIdAndAvatarByte struct {
	UserId int    `json:"user_id"`
	Avatar []byte `json:"avatar"`
}
