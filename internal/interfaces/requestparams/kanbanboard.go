package requestparams

type CreateKanbanBoardParam struct {
	Name      string `json:"name" valid:"required"`
	ProjectId int    `json:"project_id" valid:"required"`
}

type EditKanbanBoardParam struct {
	Id        int    `json:"id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	ProjectId int    `json:"project_id" valid:"required"`
}

type GetKanbanBoardsParam struct {
	ProjectId int `json:"project_id" valid:"required"`
}

type RemoveKanbanBoardParam struct {
	Id        int `json:"id" valid:"required"`
	ProjectId int `json:"project_id" valid:"required"`
}
