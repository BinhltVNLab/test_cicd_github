package requestparams

import m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"

// CreateProjectParams : params struct for creating project
type CreateProjectParams struct {
	ProjectName        string `json:"project_name" form:"project_name" valid:"required"`
	ProjectDescription string `json:"project_description" form:"project_description"`
	ManagedBy          int    `json:"managed_by" valid:"required"`
}

// ProjectIDParam : id for get and delete project
type ProjectIDParam struct {
	ProjectID int `json:"project_id" valid:"required"`
}

// UpdateProjectParams : struct of param for updaing a project
type UpdateProjectParams struct {
	ID          int        `json:"project_id" valid:"required"`
	Name        string     `json:"name" form:"project_name" valid:"required"`
	Description string     `json:"description" form:"project_description"`
	ManagedBy   int        `json:"managed_by" valid:"required"`
	Targets     []m.Target `json:"targets"`
}

// ProjectListParams : params for get list of projects
type ProjectListParams struct {
	CurrentPage int    `json:"current_page" valid:"-"`
	RowPerPage  int    `json:"row_per_page"`
	Keyword     string `json:"keyword"`
}

type ProjectIdAndNameRecord struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
