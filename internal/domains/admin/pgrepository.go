package admin

import (
	"strconv"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgAdminRepository struct {
	cm.AppRepository
}

func NewPgAdminRepository(logger echo.Logger) (repo *PgAdminRepository) {
	repo = &PgAdminRepository{}
	repo.Init(logger)
	return
}

func (repo *PgAdminRepository) InsertOrgModule(
	params *param.DataInitOrg,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		orgModule := m.OrganizationModule{
			OrganizationId: params.OrganizationId,
			Modules: params.Modules,
		}
		transErr = tx.Insert(&orgModule)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		return transErr
	})

	return err
}

func (repo *PgAdminRepository) SelectFunctionsOrg(organizationId int) ([]param.FunctionRecord, error) {
	var functionRecord []param.FunctionRecord

	q := "SELECT f.id, f.name " +
		"FROM organization_modules orgm, JSON_ARRAY_ELEMENTS(orgm.modules::json) orgms " +
		"JOIN modules as m ON (orgms->>'module_id')::int = m.id " +
		"JOIN functions AS f ON m.id = f.module_id " +
		"WHERE (orgms->>'status'='true') and ((orgm.organization_id = " + strconv.Itoa(organizationId) + ")) " +
		"AND orgm.deleted_at IS NULL " +
		"ORDER BY f.id"

	_, err := repo.DB.Query(&functionRecord, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return functionRecord, err
}

func (repo *PgAdminRepository) InsertUserPermission(
	organizationId int,
	userID int,
	functionId int,
	status int,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		userPermission := m.UserPermission{
			OrganizationId: organizationId,
			UserID: userID,
			FunctionID: functionId,
			Status: status,
		}
		transErr = tx.Insert(&userPermission)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		return transErr
	})

	return err
}

// SelectModules : get organization by tag
// Returns             : information moudles(Object)
func (repo *PgAdminRepository) SelectModules() ([]m.Module, error) {
	moudles := []m.Module{}
	err := repo.DB.Model(&moudles).
		Column("id", "name").
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return moudles, err
}

func (repo *PgAdminRepository) RemoveUserPermission(organizationId int) error {
	_, err := repo.DB.Model(&m.UserPermission{}).
		Where("organization_id = ?", organizationId).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}