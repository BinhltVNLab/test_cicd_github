package utils

import (
	valid "github.com/asaskevich/govalidator"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
)

// ValidatorFormatDisplayDate : check the field type date is correct format 2006/01/02 (YYYY/mm/dd)
// Params              :
// Returns             : return govalidator.Validator
func ValidatorFormatDisplayDate() valid.Validator {
	return valid.Validator(func(str string) bool {
		return valid.IsTime(str, cf.FormatDateDisplay)
	})
}
