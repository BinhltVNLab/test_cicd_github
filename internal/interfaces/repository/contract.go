package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type ContractRepository interface {
	SelectContractCurrentList(organizationID int, params *param.GetContractCurrentListParams) ([]param.ContractCurrentListRecord, int, error)
	SelectContractByUser(params *param.GetContractsByUserParams) ([]param.ContractByUserRecord, int, error)
	InsertContract(organizationID int,fileName string, params *param.CreateContractParams) (error)
	InsertContractType(createContractType *[]param.CreateContractType) error
	ListContractType(params *param.GetContractTypeListParams) ([]*m.ContractType, int, error)
	FindContractTypeByID(int) (*m.ContractType, error)
	UpdateContractType(*m.ContractType) error
	DeleteContractType(ID[] int) (error)
	DeleteContract(ID int) (error)
}
