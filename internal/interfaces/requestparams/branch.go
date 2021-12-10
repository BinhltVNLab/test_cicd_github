package requestparams

import "time"

type CreateBranchParam struct {
	Name     string `json:"name" valid:"required"`
	Priority int    `json:"priority" valid:"required"`
}

type EditBranchParam struct {
	Id   int    `json:"id" valid:"required"`
	Name string `json:"name" valid:"required"`
}

type SelectBranchRecords struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type RemoveBranchParam struct {
	Id int `json:"id" valid:"required"`
}

type BranchStatisticDetail struct {
	UserId            int       `json:"user_id"`
	FullName          string    `json:"full_name"`
	Birthday          time.Time `json:"birthday"`
	CompanyJoinedDate time.Time `json:"company_joined_date"`
	JobTitle          string    `json:"job_title"`
}

type BranchStatisticDetailParams struct {
	Branch      string `json:"branch" valid:"required"`
	CurrentPage int    `json:"current_page" valid:"required"`
	RowPerPage  int    `json:"row_per_page" valid:"required"`
}

type SortPrioritizationParam struct {
	Id       int `json:"id" valid:"required"`
	Priority int `json:"priority" valid:"required"`
}

type SortPrioritizationParams struct {
	OrganizationId int                       `json:"organization_id" valid:"required"`
	Priorities     []SortPrioritizationParam `json:"priorities" valid:"required"`
}
