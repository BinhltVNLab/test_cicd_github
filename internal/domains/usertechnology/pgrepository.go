package usertechnology

import (
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgUserTechnologyRepository struct {
	cm.AppRepository
}

// NewPgUserTechnologyRepository : Init PgUserTechnologyRepository
func NewPgUserTechnologyRepository(logger echo.Logger) (repo *PgUserTechnologyRepository) {
	repo = &PgUserTechnologyRepository{}
	repo.Init(logger)
	return
}

// InsertUserTechnology : Insert user technology
func (repo *PgUserTechnologyRepository) InsertUserTechnology(userID int, technologyId int) error {
	userTechnology := m.UserTechnology{
		UserID:       userID,
		TechnologyId: technologyId,
	}

	err := repo.DB.Insert(&userTechnology)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// SelectUserTechnologies : get user technologies
func (repo *PgUserTechnologyRepository) SelectTechnologiesByUserId(organizationId int, userId int) ([]param.UserTechnologyRecord, error) {
	var records []param.UserTechnologyRecord
	err := repo.DB.Model(&m.UserTechnology{}).
		ColumnExpr("user_technology.id, t.name AS technology_name").
		Join("JOIN technologies AS t ON t.id = user_technology.technology_id").
		Join("JOIN users AS u ON u.id = user_technology.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_technology.user_id = ?", userId).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

// DeleteUserTechnologyById : delete user technology
func (repo *PgUserTechnologyRepository) DeleteUserTechnologyById(organizationId int, id int) error {
	_, err := repo.DB.Model(&m.UserTechnology{}).
		TableExpr("users AS u").
		Where("u.id = user_technology.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_technology.id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgUserTechnologyRepository) DeleteUserTechnologyByTechnologyId(organizationId int, technologyId int) error {
	_, err := repo.DB.Model(&m.UserTechnology{}).
		TableExpr("users AS u").
		Where("u.id = user_technology.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_technology.technology_id = ?", technologyId).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgUserTechnologyRepository) CheckExistUserTechnology(organizationId int, userId int, technologyId int) (int, error) {
	count, err := repo.DB.Model(&m.UserTechnology{}).
		Join("JOIN users AS u ON u.id = user_technology.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_technology.user_id = ?", userId).
		Where("user_technology.technology_id = ?", technologyId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgUserTechnologyRepository) CheckExistUserTechnologyById(organizationId int, id int) (int, error) {
	count, err := repo.DB.Model(&m.UserTechnology{}).
		Join("JOIN users AS u ON u.id = user_technology.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_technology.id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

// CountUserInterestTechnology : Count user interest technology
func (repo *PgUserTechnologyRepository) CountUserInterestTechnology(organizationID int) ([]param.NumberPeopleInterestTechnology, error) {
	var records []param.NumberPeopleInterestTechnology
	err := repo.DB.Model(&m.UserTechnology{}).
		ColumnExpr("t.name AS technology").
		ColumnExpr("COUNT(t.name) AS amount").
		Join("JOIN users AS u ON u.id = user_technology.user_id").
		Join("JOIN technologies AS t ON t.id = user_technology.technology_id").
		Where("u.organization_id = ?", organizationID).
		Group("t.name", "t.priority").
		Order("t.priority").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserTechnologyRepository) SelectUsersByTechnologyName(
	organizationID int,
	technologyStatisticDetailParams *param.TechnologyStatisticDetailParams,
) ([]param.FullStatisticDetail, int, error) {
	var records []param.FullStatisticDetail
	totalRow, err := repo.DB.Model(&m.UserTechnology{}).
		Column("up.user_id", "up.birthday", "up.company_joined_date").
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		ColumnExpr("jt.name AS job_title").
		ColumnExpr("b.name AS branch").
		Join("JOIN users AS u ON u.id = user_technology.user_id").
		Join("JOIN user_profiles AS up ON up.user_id = user_technology.user_id").
		Join("LEFT JOIN job_titles AS jt ON jt.id = up.job_title").
		Join("LEFT JOIN branches AS b ON b.id = up.branch").
		Join("JOIN technologies AS t ON t.id = user_technology.technology_id").
		Where("u.organization_id = ?", organizationID).
		Where("t.name = ?", technologyStatisticDetailParams.Technology).
		Offset((technologyStatisticDetailParams.CurrentPage - 1) * technologyStatisticDetailParams.RowPerPage).
		Order("up.birthday ASC").
		Limit(technologyStatisticDetailParams.RowPerPage).
		SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}
