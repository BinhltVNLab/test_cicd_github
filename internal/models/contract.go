package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

type Contract struct {
	cm.BaseModel

	TableName            struct{} `sql:"alias:contract"`
	UserId               int
	OrganizationId       int
	ContractTypeId       int
	InsuranceSalary      int
	TotalSalary          string
	ContractStartDate    time.Time
	ContractEndDate      time.Time
	CurrencyUnit         int
	FileName             string
	LaborContractNumber  *string
	ContractCreationDate *time.Time
}

type ContractType struct {
	cm.BaseModel

	TableName        struct{} `sql:"alias:ctt"`
	OrganizationId   int
	Name             string
	FileTemplateName string
}
