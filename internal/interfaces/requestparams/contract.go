package requestparams

type GetContractCurrentListParams struct {
	UserName          string `json:"user_name"`
	ContractTypeID    int    `json:"contract_type_id"`
	BranchID          int    `json:"branch_id"`
	ContractStartDate string `json:"contract_start_date"`
	CurrentPage       int    `json:"current_page"`
	RowPerPage        int    `json:"row_per_page"`
}

type ContractCurrentListRecord struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Avatar            string `json:"avatar"`
	BranchID          int    `json:"branch_id"`
	ContractTypeID    int    `json:"contract_type_id"`
	ContractTypeName  string `json:"contract_type_name"`
	ContractStartDate string `json:"contract_start_date"`
	ContractEndDate   string `json:"contract_end_date"`
	FileTemplateName  string `json:"file_template_name"`
	InsuranceSalary   string `json:"insurance_salary"`
	TotalSalary       string `json:"total_salary"`
	CurrencyUnit      string `json:"currency_unit"`
	CompanyJoinedDate string `json:"company_joined_date"`
}

type GetContractsByUserParams struct {
	UserID      int `json:"user_id"`
	CurrentPage int `json:"current_page"`
	RowPerPage  int `json:"row_per_page"`
}

type ContractByUserRecord struct {
	ID               int    `json:"id"`
	ContractTypeID   int    `json:"contract_type_id"`
	ContractTypeName string `json:"contract_type_name"`
	FileName         string `json:"contract_file_name"`
	FileContent      string `json:"contract_file_content"`
	// FileTemplateName  	 string `json:"contract_type_file_name"`
	// FileTemplateContent  string `json:"contract_type_file_content"`
	ContractStartDate string `json:"contract_start_date"`
	ContractEndDate   string `json:"contract_end_date"`
	InsuranceSalary   string `json:"insurance_salary"`
	TotalSalary       string `json:"total_salary"`
	CurrencyUnit      string `json:"currency_unit"`
}

type CreateContractParams struct {
	UserID               int    `json:"user_id" valid:"required"`
	ContractTypeID       int    `json:"contract_type_id" valid:"required"`
	InsuranceSalary      int    `json:"insurance_salary"`
	TotalSalary          string    `json:"total_salary"`
	ContractStartDate    string `json:"contract_start_date" valid:"required"`
	ContractEndDate      string `json:"contract_end_date"`
	CurrencyUnit         int    `json:"currency_unit"`
	ContractContent      string `json:"file_content"`
	FileName             string `json:"file_name"`
	LaborContractNumber  string `json:"labor_contract_number"`
	ContractCreationDate string `json:"contract_creation_date"`
}
type CreateContractTypeParams struct {
	CreateContractType []CreateContractType `json:"contract_type_params"`
}
type CreateContractType struct {
	Name           string `json:"name" valid:"required"`
	OrganizationID int    `json:"organization_id" valid:"required,numeric"`
	TemplateFile   string `json:"template_file" valid:"required"`
	FileContent    string `json:"file_content" valid:"required"`
}

type GetContractTypeByIDParams struct {
	ContractTypeID int `json:"contract_type_id" valid:"required,numeric"`
}

type GetContractTypeListParams struct {
	OrganizationID   int    `json:"organization_id" valid:"required,numeric"`
	ContractTypeName string `json:"contract_type_name"`
	CurrentPage      int    `json:"current_page" valid:"numeric"`
	RowPerPage       int    `json:"row_per_page" valid:"range(0|100)"` // We should limit maximum of row_per_page, avoid to page 1 and row_per_page 1000000
}

type UpdateContractTypeParams struct {
	ContractTypeID   int    `json:"contract_type_id" valid:"required,numeric"`
	ContractTypeName string `json:"name" valid:"required"`
	OrganizationID   int    `json:"organization_id" valid:"required,numeric"`
	TemplateFile     string `json:"template_file"` // if wanna update template file, set this field value is not empty
}

type DeleteContractTypeParams struct {
	ID []int `json:"id" valid:"required"`
}

type DeleteContractParams struct {
	ID int `json:"contract_id" valid:"required"`
}

type PreviewContractParams struct {
	FileName    string `json:"file_name" valid:"required"`
	FileContent string `json:"file_content" valid:"required"`
}

type DeletePreviewContractParams struct {
	FileName string `json:"file_name" valid:"required"`
}
