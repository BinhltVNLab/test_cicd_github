package requestparams

type CreateKanbanListParam struct {
	ProjectId       int    `json:"project_id" valid:"required"`
	Name            string `json:"name" valid:"required"`
	KanbanBoardId   int    `json:"kanban_board_id" valid:"required"`
	PositionInBoard int    `json:"position_in_board" valid:"required"`
}

type EditKanbanListParam struct {
	ProjectId            int                   `json:"project_id" valid:"required"`
	Id                   int                   `json:"id"`
	Name                 string                `json:"name"`
	KanbanBoardId        int                   `json:"kanban_board_id" valid:"required"`
	SortPositionsInBoard []SortPositionInBoard `json:"sort_positions_in_board"`
}

type GetKanbanListsParam struct {
	KanbanBoardId int `json:"kanban_board_id" valid:"required"`
}

type RemoveKanbanListParam struct {
	Id            int `json:"id" valid:"required"`
	KanbanBoardId int `json:"kanban_board_id" valid:"required"`
	ProjectId     int `json:"project_id" valid:"required"`
}

type SortPositionInBoard struct {
	Id              int `json:"id" valid:"required"`
	PositionInBoard int `json:"position_in_board" valid:"required"`
}
