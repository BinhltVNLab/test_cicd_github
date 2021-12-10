package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
)

type BranchRepository interface {
	InsertBranch(organizationId int, params *param.CreateBranchParam, settingStep int, orgRepo OrgRepository) error
	UpdateBranch(organizationId int, id int, name string) error
	CheckExistBranch(organizationId int, id int) (int, error)
	SelectBranches(organizationId int) ([]param.SelectBranchRecords, error)
	DeleteBranch(organizationId int, id int) error
	CheckExistBranchByName(organizationId int, name string) (int, error)
	SelectUsersByBranchName(
		organizationID int,
		branchStatisticDetailParams *param.BranchStatisticDetailParams,
	) ([]param.BranchStatisticDetail, int, error)
	UpdatePriorityWithTx(params []param.SortPrioritizationParam) error
	CountBranches(organizationId int) (int, error)
}
