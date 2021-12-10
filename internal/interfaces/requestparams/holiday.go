package requestparams

type CreateHolidayParams struct {
	HolidayDate string `json:"holiday_date" valid:"required"`
	Description string `json:"description"`
}

type EditHolidayParams struct {
	Id          int    `json:"id" valid:"required"`
	HolidayDate string `json:"holiday_date"`
	Description string `json:"description"`
}

type GetHolidaysParams struct {
	Year int `json:"year" valid:"required"`
}

type RemoveHolidayParam struct {
	Id int `json:"id" valid:"required"`
}
