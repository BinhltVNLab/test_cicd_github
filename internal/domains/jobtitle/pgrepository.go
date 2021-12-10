package jobtitle

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgJobTitleRepository struct {
	cm.AppRepository
}

func NewPgJobTitleRepository(logger echo.Logger) (repo *PgJobTitleRepository) {
	repo = &PgJobTitleRepository{}
	repo.Init(logger)
	return
}

func (repo *PgJobTitleRepository) InsertJobTitle(
	organizationId int,
	params *param.CreateJobTitleParam,
	settingStep int,
	orgRepo rp.OrgRepository,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		jobTitle := m.JobTitle{
			Name:           params.Name,
			OrganizationId: organizationId,
			Priority:       params.Priority,
		}

		transErr = tx.Insert(&jobTitle)
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
		} else if settingStep < cf.JOBTITLESETTING {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.JOBTITLESETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		return transErr
	})

	return err
}

func (repo *PgJobTitleRepository) UpdateJobTitle(organizationId int, id int, name string) error {
	_, err := repo.DB.Model(&m.JobTitle{Name: name}).
		Column("name", "updated_at").
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgJobTitleRepository) CheckExistJobTitle(organizationId int, id int) (int, error) {
	count, err := repo.DB.Model(&m.JobTitle{}).
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgJobTitleRepository) CheckExistJobTitleByName(organizationId int, name string) (int, error) {
	count, err := repo.DB.Model(&m.JobTitle{}).
		Where("name = ?", name).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgJobTitleRepository) SelectJobTitles(organizationId int) ([]param.SelectJobTitleRecords, error) {
	var records []param.SelectJobTitleRecords
	err := repo.DB.Model(&m.JobTitle{}).
		Column("id", "name").
		Where("organization_id = ?", organizationId).
		Order("priority ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgJobTitleRepository) CountJobTitles(organizationId int) (int, error) {
	count, err := repo.DB.Model(&m.JobTitle{}).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgJobTitleRepository) DeleteJobTitle(organizationId int, id int) error {
	_, err := repo.DB.Model(&m.JobTitle{}).
		Where("organization_id = ?", organizationId).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgJobTitleRepository) SelectUsersByJobTitleName(
	organizationID int,
	jobTitleStatisticDetailParams *param.JobTitleStatisticDetailParams,
) ([]param.JobTitleStatisticDetail, int, error) {
	var records []param.JobTitleStatisticDetail
	totalRow, err := repo.DB.Model(&m.UserProfile{}).
		Column("user_profile.user_id", "user_profile.birthday", "user_profile.company_joined_date").
		ColumnExpr("user_profile.first_name || ' ' || user_profile.last_name full_name").
		ColumnExpr("b.name AS branch").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Join("JOIN job_titles AS jt ON jt.id = user_profile.job_title").
		Join("LEFT JOIN branches AS b ON b.id = user_profile.branch").
		Where("u.organization_id = ?", organizationID).
		Where("jt.name = ?", jobTitleStatisticDetailParams.JobTitle).
		Offset((jobTitleStatisticDetailParams.CurrentPage - 1) * jobTitleStatisticDetailParams.RowPerPage).
		Order("user_profile.birthday ASC").
		Limit(jobTitleStatisticDetailParams.RowPerPage).
		SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgJobTitleRepository) UpdatePriorityWithTx(params []param.SortPrioritizationParam) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		if len(params) > 0 {
			for _, pr := range params {
				_, errTx := tx.Model(&m.JobTitle{Priority: pr.Priority}).
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
