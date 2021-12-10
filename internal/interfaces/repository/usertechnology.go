package repository

import param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"

type UserTechnologyRepository interface {
	InsertUserTechnology(userID int, technologyId int) error
	CheckExistUserTechnology(organizationId int, userId int, technologyId int) (int, error)
	CheckExistUserTechnologyById(organizationId int, id int) (int, error)
	SelectTechnologiesByUserId(organizationId int, userId int) ([]param.UserTechnologyRecord, error)
	DeleteUserTechnologyById(organizationId int, id int) error
	DeleteUserTechnologyByTechnologyId(organizationId int, technologyId int) error
	CountUserInterestTechnology(organizationID int) ([]param.NumberPeopleInterestTechnology, error)
	SelectUsersByTechnologyName(
		organizationID int,
		technologyStatisticDetailParams *param.TechnologyStatisticDetailParams,
	) ([]param.FullStatisticDetail, int, error)
}
