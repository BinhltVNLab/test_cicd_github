package requestparams

import "time"

type CreateJobTitleParam struct {
	Name     string `json:"name" valid:"required"`
	Priority int    `json:"priority" valid:"required"`
}

type EditJobTitleParam struct {
	Id   int    `json:"id" valid:"required"`
	Name string `json:"name" valid:"required"`
}

type SelectJobTitleRecords struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type RemoveJobTitleParam struct {
	Id int `json:"id" valid:"required"`
}

type JobTitleStatisticDetail struct {
	UserId            int       `json:"user_id"`
	FullName          string    `json:"full_name"`
	Birthday          time.Time `json:"birthday"`
	CompanyJoinedDate time.Time `json:"company_joined_date"`
	Branch            string    `json:"branch"`
}

type JobTitleStatisticDetailParams struct {
	JobTitle    string `json:"job_title" valid:"required"`
	CurrentPage int    `json:"current_page" valid:"required"`
	RowPerPage  int    `json:"row_per_page" valid:"required"`
}
