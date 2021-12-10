package repository

import param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"

type JobTitleRepository interface {
	InsertJobTitle(
		organizationId int,
		params *param.CreateJobTitleParam,
		settingStep int,
		orgRepo OrgRepository,
	) error
	UpdateJobTitle(organizationId int, id int, name string) error
	CheckExistJobTitle(organizationId int, id int) (int, error)
	SelectJobTitles(organizationId int) ([]param.SelectJobTitleRecords, error)
	DeleteJobTitle(organizationId int, id int) error
	CheckExistJobTitleByName(organizationId int, name string) (int, error)
	SelectUsersByJobTitleName(
		organizationID int,
		jobTitleStatisticDetailParams *param.JobTitleStatisticDetailParams,
	) ([]param.JobTitleStatisticDetail, int, error)
	UpdatePriorityWithTx(params []param.SortPrioritizationParam) error
	CountJobTitles(organizationId int) (int, error)
}
