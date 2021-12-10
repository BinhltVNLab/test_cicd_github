package requestparams

type CreateTechnologyParam struct {
	Name     string `json:"name" valid:"required"`
	Priority int    `json:"priority" valid:"required"`
}

type EditTechnologyParam struct {
	Id   int    `json:"id" valid:"required"`
	Name string `json:"name" valid:"required"`
}

type SelectTechnologyRecords struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type RemoveTechnologyParam struct {
	Id int `json:"id" valid:"required"`
}
