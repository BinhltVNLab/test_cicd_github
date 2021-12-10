package technology

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgTechnologyRepository struct {
	cm.AppRepository
}

func NewPgTechnologyRepository(logger echo.Logger) (repo *PgTechnologyRepository) {
	repo = &PgTechnologyRepository{}
	repo.Init(logger)
	return
}

func (repo *PgTechnologyRepository) InsertTechnology(
	organizationId int,
	params *param.CreateTechnologyParam,
	settingStep int,
	orgRepo rp.OrgRepository,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		technology := m.Technology{
			Name:           params.Name,
			OrganizationId: organizationId,
			Priority:       params.Priority,
		}

		transErr = tx.Insert(&technology)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		if settingStep == 0 {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.ORGANIZATIONEMAILSETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		} else if settingStep < cf.TECHNOLOGYSETTING {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.TECHNOLOGYSETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		return transErr
	})

	return err
}

func (repo *PgTechnologyRepository) UpdateTechnology(organizationId int, id int, name string) error {
	_, err := repo.DB.Model(&m.Technology{Name: name}).
		Column("name", "updated_at").
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgTechnologyRepository) CheckExistTechnology(organizationId int, id int) (int, error) {
	count, err := repo.DB.Model(&m.Technology{}).
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgTechnologyRepository) CheckExistTechnologyByName(organizationId int, name string) (int, error) {
	count, err := repo.DB.Model(&m.Technology{}).
		Where("name = ?", name).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgTechnologyRepository) SelectTechnologies(organizationId int) ([]param.SelectTechnologyRecords, error) {
	var records []param.SelectTechnologyRecords
	err := repo.DB.Model(&m.Technology{}).
		Column("id", "name").
		Where("organization_id = ?", organizationId).
		Order("priority ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgTechnologyRepository) CountTechnologies(organizationId int) (int, error) {
	count, err := repo.DB.Model(&m.Technology{}).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgTechnologyRepository) DeleteTechnology(organizationId int, id int) error {
	_, err := repo.DB.Model(&m.Technology{}).
		Where("organization_id = ?", organizationId).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgTechnologyRepository) UpdatePriorityWithTx(params []param.SortPrioritizationParam) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		if len(params) > 0 {
			for _, pr := range params {
				_, errTx := tx.Model(&m.Technology{Priority: pr.Priority}).
					Column("priority", "updated_at").
					Where("id = ?", pr.Id).
					Update()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}
			}
		}

		return errTx
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
