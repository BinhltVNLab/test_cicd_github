package requestparams

import "time"

type CreateUserTechnologyParam struct {
	UserId       int `json:"user_id" valid:"required"`
	TechnologyId int `json:"technology_id" valid:"required"`
}

type GetTechnologiesOfUserParam struct {
	UserId int `json:"user_id" valid:"required"`
}

type RemoveTechnologyOfUserParam struct {
	Id     int `json:"id" valid:"required"`
	UserId int `json:"user_id" valid:"required"`
}

type UserTechnologyRecord struct {
	Id             int    `json:"id"`
	TechnologyName string `json:"technology_name"`
}

// NumberPeopleInterestTechnology : struct for all people interest technology
type NumberPeopleInterestTechnology struct {
	Technology string `json:"technology"`
	Amount     int    `json:"amount"`
}

type FullStatisticDetail struct {
	UserId            int       `json:"user_id"`
	FullName          string    `json:"full_name"`
	JobTitle          string    `json:"job_title"`
	Birthday          time.Time `json:"birthday"`
	CompanyJoinedDate time.Time `json:"company_joined_date"`
	Branch            string    `json:"branch"`
}

type TechnologyStatisticDetailParams struct {
	Technology  string `json:"technology" valid:"required"`
	CurrentPage int    `json:"current_page" valid:"required"`
	RowPerPage  int    `json:"row_per_page" valid:"required"`
}
