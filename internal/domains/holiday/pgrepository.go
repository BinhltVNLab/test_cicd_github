package holiday

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type PgHolidayRepository struct {
	cm.AppRepository
}

func NewPgHolidayRepository(logger echo.Logger) (repo *PgHolidayRepository) {
	repo = &PgHolidayRepository{}
	repo.Init(logger)
	return
}

func (repo *PgHolidayRepository) InsertHoliday(
	organizationId int,
	params *param.CreateHolidayParams,
	settingStep int,
	orgRepo rp.OrgRepository,
) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		holiday := m.Holiday{
			OrganizationId: organizationId,
			HolidayDate:    calendar.ParseTime(cf.FormatDateDatabase, params.HolidayDate),
			Description:    params.Description,
		}

		transErr = tx.Insert(&holiday)
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
		} else if settingStep < cf.FINISHSETTING {
			transErr = orgRepo.UpdateSettingStepWithTx(tx, organizationId, cf.FINISHSETTING)
			if transErr != nil {
				repo.Logger.Error(transErr)
				return transErr
			}
		}

		return transErr
	})

	return err
}

func (repo *PgHolidayRepository) UpdateHoliday(params *param.EditHolidayParams) error {
	holiday := m.Holiday{
		HolidayDate: calendar.ParseTime(cf.FormatDateDatabase, params.HolidayDate),
		Description: params.Description,
	}

	_, err := repo.DB.Model(&holiday).
		Column("holiday_date", "description", "updated_at").
		Where("id = ?", params.Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgHolidayRepository) CheckHolidayExistByDate(holidayDate string, organizationId int) (int, error) {
	count, err := repo.DB.Model(&m.Holiday{}).
		Where("organization_id = ?", organizationId).
		Where("holiday_date = ?::date", holidayDate).
		Count()
	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgHolidayRepository) CheckHolidayExistById(id int) (int, error) {
	count, err := repo.DB.Model(&m.Holiday{}).Where("id = ?", id).Count()
	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgHolidayRepository) SelectHolidays(organizationId int, year int, columns ...string) ([]m.Holiday, error) {
	var holidays []m.Holiday
	err := repo.DB.Model(&holidays).
		Column(columns...).
		Where("organization_id = ?", organizationId).
		Where("EXTRACT(YEAR FROM holiday_date) = ?", year).
		Order("holiday_date ASC").
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return holidays, err
}

func (repo *PgHolidayRepository) DeleteHoliday(id int) error {
	_, err := repo.DB.Model(&m.Holiday{}).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
