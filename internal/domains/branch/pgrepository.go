package branch

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgBranchRepository struct {
	cm.AppRepository
}

func NewPgBranchRepository(logger echo.Logger) (repo *PgBranchRepository) {
	repo = &PgBranchRepository{}
	repo.Init(logger)
	return
}

func (repo *PgBranchRepository) InsertBranch(
	organizationId int,
	params *param.CreateBranchParam,
	settingStep int,
	orgRepo rp.OrgRepository,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		branch := m.Branch{
			Name:           params.Name,
			OrganizationId: organizationId,
			Priority:       params.Priority,
		}
		transErr = tx.Insert(&branch)
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
		} else if settingStep < cf.BRANCHSETTING {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.BRANCHSETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		return transErr
	})

	return err
}

func (repo *PgBranchRepository) UpdateBranch(organizationId int, id int, name string) error {
	_, err := repo.DB.Model(&m.Branch{Name: name}).
		Column("name", "updated_at").
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgBranchRepository) CheckExistBranch(organizationId int, id int) (int, error) {
	count, err := repo.DB.Model(&m.Branch{}).
		Where("id = ?", id).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgBranchRepository) CheckExistBranchByName(organizationId int, name string) (int, error) {
	count, err := repo.DB.Model(&m.Branch{}).
		Where("name = ?", name).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgBranchRepository) SelectBranches(organizationId int) ([]param.SelectBranchRecords, error) {
	var records []param.SelectBranchRecords
	err := repo.DB.Model(&m.Branch{}).
		Column("id", "name").
		Where("organization_id = ?", organizationId).
		Order("priority ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgBranchRepository) CountBranches(organizationId int) (int, error) {
	count, err := repo.DB.Model(&m.Branch{}).
		Where("organization_id = ?", organizationId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgBranchRepository) DeleteBranch(organizationId int, id int) error {
	_, err := repo.DB.Model(&m.Branch{}).
		Where("organization_id = ?", organizationId).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgBranchRepository) SelectUsersByBranchName(
	organizationID int,
	branchStatisticDetailParams *param.BranchStatisticDetailParams,
) ([]param.BranchStatisticDetail, int, error) {
	var records []param.BranchStatisticDetail
	totalRow, err := repo.DB.Model(&m.UserProfile{}).
		Column("user_profile.user_id", "user_profile.birthday", "user_profile.company_joined_date").
		ColumnExpr("user_profile.first_name || ' ' || user_profile.last_name full_name").
		ColumnExpr("jt.name AS job_title").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Join("LEFT JOIN job_titles AS jt ON jt.id = user_profile.job_title").
		Join("JOIN branches AS b ON b.id = user_profile.branch").
		Where("u.organization_id = ?", organizationID).
		Where("b.name = ?", branchStatisticDetailParams.Branch).
		Offset((branchStatisticDetailParams.CurrentPage - 1) * branchStatisticDetailParams.RowPerPage).
		Order("user_profile.birthday ASC").
		Limit(branchStatisticDetailParams.RowPerPage).
		SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgBranchRepository) UpdatePriorityWithTx(params []param.SortPrioritizationParam) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		if len(params) > 0 {
			for _, pr := range params {
				_, errTx := tx.Model(&m.Branch{Priority: pr.Priority}).
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
