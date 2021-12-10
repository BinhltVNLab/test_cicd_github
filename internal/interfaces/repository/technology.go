package repository

import param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"

type TechnologyRepository interface {
	InsertTechnology(
		organizationId int,
		params *param.CreateTechnologyParam,
		settingStep int,
		orgRepo OrgRepository,
	) error
	UpdateTechnology(organizationId int, id int, name string) error
	CheckExistTechnology(organizationId int, id int) (int, error)
	SelectTechnologies(organizationId int) ([]param.SelectTechnologyRecords, error)
	DeleteTechnology(organizationId int, id int) error
	CheckExistTechnologyByName(organizationId int, name string) (int, error)
	UpdatePriorityWithTx(params []param.SortPrioritizationParam) error
	CountTechnologies(organizationId int) (int, error)
}
