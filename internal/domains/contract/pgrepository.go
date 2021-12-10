package contract

import (
	"time"
	"math/rand"
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/go-pg/pg/v9"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type PgContractRepository struct {
	cm.AppRepository
}

func NewPgContractRepository(logger echo.Logger) (repo *PgContractRepository) {
	repo = &PgContractRepository{}
	repo.Init(logger)
	return
}

func (repo *PgContractRepository) SelectContractCurrentList(organizationId int, getContractCurrentListParams *param.GetContractCurrentListParams) ([]param.ContractCurrentListRecord, int, error) {
	var records []param.ContractCurrentListRecord
	q := repo.DB.Model(&m.Contract{})

	q.Column("contract.id", "contract.user_id", "contract.insurance_salary", "contract.total_salary", "contract.contract_start_date", "contract.contract_end_date", "contract.currency_unit",
		"contract.contract_type_id", "up.first_name", "up.last_name", "up.avatar", "up.company_joined_date").
		ColumnExpr("ctt.name as contract_type_name").
		ColumnExpr("up.branch as branch_id").
		Join("JOIN contract_types as ctt on contract.contract_type_id = ctt.id").
		Join("FULL OUTER JOIN user_profiles as up on up.user_id = contract.user_id").
		Where("contract.organization_id = ?", organizationId).
		Where("up.status != 3")

	if getContractCurrentListParams.ContractTypeID != 0 {
		q.Where("contract.contract_type_id = ?", getContractCurrentListParams.ContractTypeID)
	}

	if getContractCurrentListParams.BranchID != 0 {
		q.Where("up.branch = ?", getContractCurrentListParams.BranchID)
	}

	if getContractCurrentListParams.ContractStartDate != "" {
		q.Where("contract.contract_start_date = ?", getContractCurrentListParams.ContractStartDate)
	}

	if getContractCurrentListParams.UserName != "" {
		userName := "%" + getContractCurrentListParams.UserName + "%"
		q.Where("vietnamese_unaccent(LOWER(up.first_name)) || ' ' || vietnamese_unaccent(LOWER(up.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	q.Offset((getContractCurrentListParams.CurrentPage - 1) * getContractCurrentListParams.RowPerPage).
		Order("contract.created_at DESC").
		Limit(getContractCurrentListParams.RowPerPage)

	totalRow, err := q.SelectAndCount(&records)
	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return records, totalRow, err
}

func (repo *PgContractRepository) SelectContractByUser(getContractByUserParams *param.GetContractsByUserParams) ([]param.ContractByUserRecord, int, error) {
	var records []param.ContractByUserRecord
	q := repo.DB.Model(&m.Contract{})
	q.Column("contract.id", "contract.insurance_salary", "contract.total_salary", "contract.contract_start_date", "contract.contract_end_date", "contract.currency_unit",
		"contract.contract_type_id", "contract.file_name", "ctt.file_template_name").
		ColumnExpr("ctt.name as contract_type_name").
		ColumnExpr("ctt.file_template_name as file_template_name").
		Join("JOIN contract_types as ctt on contract.contract_type_id = ctt.id").
		Where("contract.user_id = ?", getContractByUserParams.UserID)

	totalRow, err := q.SelectAndCount(&records)

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return records, totalRow, err
}
func RandomString(n int, fileName string) string {
    var letters = []rune(fileName)
 
    s := make([]rune, n)
    for i := range s {
        s[i] = letters[rand.Intn(len(letters))]
    }
    return string(s)
}

// InsertContract : insert new target evaluation form
// Params           : createContractParams
func (repo *PgContractRepository) InsertContract(organizationId int,fileName string, createContractParams  *param.CreateContractParams) error {
	contractStartDate := calendar.ParseTime(cf.FormatDateDatabase, createContractParams.ContractStartDate)

	var contractEndDate time.Time
	if createContractParams.ContractEndDate != "" {
		contractEndDate = calendar.ParseTime(cf.FormatDateDatabase, createContractParams.ContractEndDate)
	}

	var (
		contractCreationDate *time.Time
		laborContractNumber  *string
	)

	if createContractParams.ContractCreationDate != "" {
		t, _ := time.Parse(cf.FormatDateDatabase, createContractParams.ContractCreationDate)
		contractCreationDate = &t
	}

	if createContractParams.LaborContractNumber != "" {
		laborContractNumber = &createContractParams.LaborContractNumber
	}
	enCodeSalary := base64.StdEncoding.EncodeToString([]byte(createContractParams.TotalSalary))

	contract := m.Contract{
		UserId:            createContractParams.UserID,
		OrganizationId:    organizationId,
		ContractTypeId:    createContractParams.ContractTypeID,
		InsuranceSalary:   createContractParams.InsuranceSalary,
		TotalSalary:       RandomString(3, fileName) + enCodeSalary + RandomString(5, fileName),
		ContractStartDate: contractStartDate,
		ContractEndDate:   contractEndDate,
		CurrencyUnit:      createContractParams.CurrencyUnit,
		FileName:          fileName,
		LaborContractNumber: laborContractNumber,
		ContractCreationDate: contractCreationDate,
	}

	err := repo.DB.Insert(&contract)

	return err
}


func (repo *PgContractRepository) InsertContractType(createContractType *[]param.CreateContractType) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		for _, contractTypeParams := range *createContractType {
			contracParams := m.ContractType{
				OrganizationId : contractTypeParams.OrganizationID,
				Name: 			contractTypeParams.Name,
				FileTemplateName: contractTypeParams.TemplateFile,
			}
			errTx = tx.Insert(&contracParams)
			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}
		}

		return errTx
	})
	return err
}

func (repo *PgContractRepository) ListContractType(params *param.GetContractTypeListParams) ([]*m.ContractType, int, error) {
	var (
		contractTypes []*m.ContractType
		totalRow      int
		err           error
	)

	q := repo.DB.Model(&m.ContractType{}).Where("organization_id = ?", params.OrganizationID)
	if params.ContractTypeName != "" {
		q = q.Where("LOWER(name) LIKE LOWER(?)", "%" + params.ContractTypeName + "%")
	}

	if  totalRow, err = q.Count(); err != nil {
		return nil, 0, err
	}

	if ! (params.CurrentPage == 0 && params.RowPerPage == 0) {
		q = q.Limit(params.RowPerPage).Offset((params.CurrentPage - 1) * params.RowPerPage)
	}

	if _, err = q.Order("updated_at desc").
		SelectAndCount(&contractTypes); err != nil {
			return nil, 0, err
	}

	return contractTypes, totalRow, nil
}

func (repo *PgContractRepository) FindContractTypeByID(id int) (*m.ContractType, error) {
	var (
		contractType m.ContractType
		err          error
	)

	if err = repo.DB.Model(&contractType).Where("id = ?", id).Select(); err != nil {
		return nil, err
	}

	return &contractType, nil
}

func (repo *PgContractRepository) UpdateContractType(contractType *m.ContractType) error {
	_, err := repo.DB.Model(contractType).WherePK().Update()

	return err
}

// DeleteContractType : Remove contracttype
func (repo *PgContractRepository) DeleteContractType(ID[] int) error {
	var errTx error
	for _, Id := range ID {
		_, err := repo.DB.Model(&m.ContractType{}).
		Where("id = ?", Id).
		Delete()
		if err != nil {
			repo.Logger.Error(err)
		}
	}
	return errTx
}

// DeleteContract : Remove contract
func (repo *PgContractRepository) DeleteContract(ID int) error {
	_, err := repo.DB.Model(&m.Contract{}).
		Where("id = ?", ID).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}