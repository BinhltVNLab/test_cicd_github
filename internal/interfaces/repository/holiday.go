package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type HolidayRepository interface {
	InsertHoliday(
		organizationId int,
		params *param.CreateHolidayParams,
		settingStep int,
		orgRepo OrgRepository,
	) error
	UpdateHoliday(params *param.EditHolidayParams) error
	CheckHolidayExistByDate(holidayDate string, organizationId int) (int, error)
	CheckHolidayExistById(id int) (int, error)
	SelectHolidays(organizationId int, year int, columns ...string) ([]m.Holiday, error)
	DeleteHoliday(id int) error
}
