package contract

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
)

type Controller struct {
	cm.BaseController

	contractRepository rp.ContractRepository
	userRepo           rp.UserRepository
	BranchRepo         rp.BranchRepository
	cloud              gc.StorageUtility
}

func NewContractController(logger echo.Logger, contractRepository rp.ContractRepository, userRepo rp.UserRepository, branchRepo rp.BranchRepository, cloud gc.StorageUtility) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, contractRepository, userRepo, branchRepo, cloud}
	ctr.Init(logger)
	return
}

func (ctr *Controller) GetContractCurrentList(c echo.Context) error {
	getContractCurrentListParams := new(param.GetContractCurrentListParams)

	if err := c.Bind(getContractCurrentListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getContractCurrentListParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	contractRecords, totalRow, err := ctr.contractRepository.SelectContractCurrentList(
		userProfile.OrganizationID,
		getContractCurrentListParams,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var contractCurrentList []map[string]interface{}
	for _, contract := range contractRecords {
		var base64Img []byte = nil
		if contract.Avatar != "" {
			base64Img, err = ctr.cloud.GetFileByFileName(contract.Avatar, cf.AvatarFolderGCS)

			if err != nil {
				ctr.Logger.Error(err)
				base64Img = nil
			}
		}

		res := map[string]interface{}{
			"id":                  contract.ID,
			"user_id":             contract.UserID,
			"first_name":          contract.FirstName,
			"last_name":           contract.LastName,
			"avatar":              base64Img,
			"branch_id":           contract.BranchID,
			"contract_type_id":    contract.ContractTypeID,
			"contract_type_name":  contract.ContractTypeName,
			"contract_start_date": contract.ContractStartDate,
			"contract_end_date":   contract.ContractEndDate,
			"file_template_name":  contract.FileTemplateName,
			"insurance_salary":    contract.InsuranceSalary,
			"total_salary":        contract.TotalSalary,
			"currency_unit":       contract.CurrencyUnit,
			"company_joined_date": contract.CompanyJoinedDate,
		}
		contractCurrentList = append(contractCurrentList, res)
	}

	pagination := map[string]interface{}{
		"current_page": getContractCurrentListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getContractCurrentListParams.RowPerPage,
	}

	responseData := map[string]interface{}{
		"pagination":    pagination,
		"contract_list": contractCurrentList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get contract current list successfully.",
		Data:    responseData,
	})
}

func (ctr *Controller) GetContractByUser(c echo.Context) error {
	getContractsByUserParams := new(param.GetContractsByUserParams)

	if err := c.Bind(getContractsByUserParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getContractsByUserParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	contractRecords, totalRow, err := ctr.contractRepository.SelectContractByUser(getContractsByUserParams)

	userProfile := c.Get("user_profile").(m.User)
	var contractList []map[string]interface{}
	for _, contract := range contractRecords {
		filePath := "https://storage.googleapis.com/" + os.Getenv("GOOGLE_STORAGE_BUCKET") + "/" + cf.CONTRACTFOLDERGCS +
			strconv.Itoa(userProfile.OrganizationID) + "/" + contract.FileName
		res := map[string]interface{}{
			"id":                  contract.ID,
			"contract_type_id":    contract.ContractTypeID,
			"contract_type_name":  contract.ContractTypeName,
			"contract_file_name":  contract.FileName,
			"contract_start_date": contract.ContractStartDate,
			"contract_end_date":   contract.ContractEndDate,
			"insurance_salary":    contract.InsuranceSalary,
			"total_salary":        contract.TotalSalary,
			"currency_unit":       contract.CurrencyUnit,
			"file_path":           filePath,
		}
		contractList = append(contractList, res)
	}

	pagination := map[string]interface{}{
		"current_page": getContractsByUserParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getContractsByUserParams.RowPerPage,
	}

	responseData := map[string]interface{}{
		"pagination":    pagination,
		"contract_list": contractList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get contract by user successfully.",
		Data:    responseData,
	})
}

func (ctr *Controller) CreateContract(c echo.Context) error {
	createContractParams := new(param.CreateContractParams)
	if err := c.Bind(createContractParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(createContractParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}
	userProfile := c.Get("user_profile").(m.User)
	millisecondTimeNow := int(time.Now().UnixNano() / int64(time.Millisecond))
	
		
		if createContractParams.ContractCreationDate != "" {
			_, err := time.Parse(cf.FormatDateDatabase, createContractParams.ContractCreationDate)
			if err != nil {
				return c.JSON(http.StatusBadRequest, cf.JsonResponse{
					Status:  cf.FailResponseCode,
					Message: "Invalid value for field contract_creation_date",
				})
			}
		}
	
		err = ctr.cloud.UploadFileToCloud(
			createContractParams.ContractContent,
			strings.Replace(createContractParams.FileName, " ", "_", -1),
			cf.CONTRACTFOLDERGCS+strconv.Itoa(userProfile.OrganizationID)+"/"+
				strconv.Itoa(millisecondTimeNow)+"_",
		)
	
		if err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
		fileName := strconv.Itoa(millisecondTimeNow) + "_" + createContractParams.FileName

		err = ctr.contractRepository.InsertContract(userProfile.OrganizationID, strings.Replace(fileName, " ", "_", -1), createContractParams)
	
		if err != nil && err.Error() != pg.ErrNoRows.Error() {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "System Error",
				Data:    err,
			})
		}
	
		var contractEndDate = time.Time{}
		if createContractParams.ContractEndDate != "" {
			contractEndDateParse, _ := time.Parse(cf.FormatDateDatabase, createContractParams.ContractEndDate)
			contractEndDate = contractEndDateParse
		}
		err = ctr.userRepo.UpdateContractExpireDate(createContractParams.UserID, contractEndDate)
		
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}
	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create contract successfully.",
	})
}

func (ctr *Controller) CreateContractType(c echo.Context) error {
	params := new(param.CreateContractTypeParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if err := ctr.contractRepository.InsertContractType(&params.CreateContractType); err != nil {
		if err != nil && strings.Contains(err.Error(), "duplicate key value violates") {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Duplicate contract type value violates",
			})
		}

		ctr.Logger.Errorf("Insert new contract type to DB got err: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Failed to insert new contract type to DB",
		})
	}
	for _, item := range params.CreateContractType {
		_, err = ctr.uploadContractTypeTemplateFile(item.TemplateFile, item.FileContent)
		if err != nil {
		ctr.Logger.Errorf(fmt.Sprintf("upload file to gcp got error: %v", err))
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Upload file to GCP got error",
		})
	}
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create new contract type successful",
	})
}

// Get contract type list
func (ctr *Controller) GetContractTypeList(c echo.Context) error {
	getContractTypeListParams := new(param.GetContractTypeListParams)

	if err := c.Bind(getContractTypeListParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(getContractTypeListParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	if getContractTypeListParams.CurrentPage < 0 {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "The current_page must be >= 0",
		})
	}

	contractRecords, totalRow, err := ctr.contractRepository.ListContractType(getContractTypeListParams)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	pagination := map[string]interface{}{
		"current_page": getContractTypeListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": getContractTypeListParams.RowPerPage,
	}

	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	contractTypes := make([]map[string]interface{}, 0, len(contractRecords))
	for _, row := range contractRecords {
		data := map[string]interface{}{
			"contract_type_id":   row.ID,
			"name":               row.Name,
			"file_template_name": row.FileTemplateName,
			"created_at":         row.CreatedAt.In(location).Format(cf.FormatDateDisplay),
			"updated_at":         row.UpdatedAt.In(location).Format(cf.FormatDateDisplay),
			"deleted_at":         nil,
		}
		var filePath = ""
		if row.FileTemplateName != "" {
			filePath = "https://storage.googleapis.com/" + os.Getenv("GOOGLE_STORAGE_BUCKET") + "/" + cf.ContractTypeFolderGCS + row.FileTemplateName
		}
		data["file_template_content"] = filePath

		contractTypes = append(contractTypes, data)
	}

	users, err := ctr.userRepo.GetAllUserNameByOrgID(getContractTypeListParams.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get user list failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	userList := make(map[int]interface{})
	for i := 0; i < len(users); i++ {
		userList[users[i].UserID] = users[i].FullName
	}

	responseData := map[string]interface{}{
		"pagination":     pagination,
		"contract_types": contractTypes,
		"users":          userList,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    responseData,
	})
}

func (ctr *Controller) EditContractType(c echo.Context) error {
	params := new(param.UpdateContractTypeParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	var (
		contractType *m.ContractType
		err          error
	)

	if contractType, err = ctr.contractRepository.FindContractTypeByID(params.ContractTypeID); err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Contract type is not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: fmt.Sprintf("Query contract type by id=%v got error: %v", params.ContractTypeID, err),
		})
	}

	contractType.Name = params.ContractTypeName
	contractType.OrganizationId = params.OrganizationID

	if params.TemplateFile != "" {
		extension, err := ctr.validateContractTypeTemplateFile(params.TemplateFile)
		if err != nil {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Invalid Contract Type File Template",
			})
		}
		fileName, err := ctr.uploadContractTypeEditTemplateFile(contractType.Name, params.TemplateFile, extension)
		if err != nil {
			ctr.Logger.Errorf(fmt.Sprintf("upload file to gcp got error: %v", err))
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Upload file to GCP got error",
			})
		}

		// Delete old contract template file template on GCS
		if err := ctr.cloud.DeleteFileCloud(contractType.FileTemplateName, cf.ContractTypeFolderGCS); err != nil {
			return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Failed to delete old contract type template on GCS",
			})
		}

		contractType.FileTemplateName = fileName
	}

	if err := ctr.contractRepository.UpdateContractType(contractType); err != nil {
		ctr.Logger.Errorf("Update contract type to DB got err: %v", err)
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Failed to update contract type to DB",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Contract Type Updated Successful",
		Data: map[string]interface{}{
			"contract_type_id":   contractType.ID,
			"organization_id":    contractType.OrganizationId,
			"contact_type_name":  contractType.Name,
			"file_template_name": contractType.FileTemplateName,
		},
	})
}

// Delete contract type by ID
func (ctr *Controller) DeleteContractTypeByID(c echo.Context) error {
	deleteContractTypeParams := new(param.DeleteContractTypeParams)

	if err := c.Bind(deleteContractTypeParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(deleteContractTypeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.contractRepository.DeleteContractType(
		deleteContractTypeParams.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove contract type successfully.",
	})
}

// Delete contract by ID
func (ctr *Controller) DeleteContractByID(c echo.Context) error {
	deleteContractParams := new(param.DeleteContractParams)

	if err := c.Bind(deleteContractParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(deleteContractParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	err = ctr.contractRepository.DeleteContract(
		deleteContractParams.ID,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove contract successfully.",
	})
}

func (ctr *Controller) PreviewContract(c echo.Context) error {
	previewContractParams := new(param.PreviewContractParams)

	if err := c.Bind(previewContractParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(previewContractParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err = ctr.cloud.UploadFileToCloud(
		previewContractParams.FileContent,
		previewContractParams.FileName,
		cf.CONTRACTFOLDERGCS+strconv.Itoa(userProfile.OrganizationID)+"/",
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	filePath := "https://storage.googleapis.com/" + os.Getenv("GOOGLE_STORAGE_BUCKET") + "/" + cf.CONTRACTFOLDERGCS +
		strconv.Itoa(userProfile.OrganizationID) + "/" + previewContractParams.FileName

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create temporary file successfully.",
		Data:    filePath,
	})
}

func (ctr *Controller) DeletePreviewContract(c echo.Context) error {
	deletePreviewContractParams := new(param.DeletePreviewContractParams)

	if err := c.Bind(deletePreviewContractParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	_, err := valid.ValidateStruct(deletePreviewContractParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err = ctr.cloud.DeleteFileCloud(
		deletePreviewContractParams.FileName,
		cf.CONTRACTFOLDERGCS+strconv.Itoa(userProfile.OrganizationID)+"/",
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Delete temporary file successfully.",
	})
}

func (ctr *Controller) GetContractTypeByID(c echo.Context) error {
	params := new(param.GetContractTypeByIDParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	var (
		contractType *m.ContractType
		err          error
	)

	if contractType, err = ctr.contractRepository.FindContractTypeByID(params.ContractTypeID); err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusBadRequest, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Contract type is not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: fmt.Sprintf("Query contract type by id=%v got error: %v", params.ContractTypeID, err),
		})
	}

	base64ContentFile, err := ctr.cloud.GetFileByFileName(contractType.FileTemplateName, cf.ContractTypeFolderGCS)

	if err != nil {
		ctr.Logger.Error(err)
		base64ContentFile = nil
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get contract type successful",
		Data:    base64ContentFile,
	})
}
func (ctr *Controller) validateContractTypeTemplateFile(content string) (string, error) {
	// Validate template file type & file size
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content))
	buf := &bytes.Buffer{}
	size, err := io.Copy(buf, dec)
	if err != nil {
		return "", fmt.Errorf("validate template file got error %v", err)
	}

	if size/1024/1024 > cf.MaxTemplateFileSize {
		return "", fmt.Errorf("the template file size=%v KB is invalid", size)
	}

	fType := http.DetectContentType(buf.Bytes())
	extension := ""

	for mType, e := range cf.TemplateFileType {
		if fType == mType {
			extension = e
		}
	}

	if extension == "" {
		return "", fmt.Errorf("the template file type is invalid")
	}

	return extension, nil
}

func (ctr *Controller) uploadContractTypeTemplateFile(fileName, content string) (string, error) {
	if err := ctr.cloud.UploadFileToCloud(content, fileName, cf.ContractTypeFolderGCS); err != nil {
		return "", fmt.Errorf("upload template file to GCS got err: %v", err)
	}

	return fileName, nil
}
	func (ctr *Controller) uploadContractTypeEditTemplateFile(name, content, extension string) (string, error) {

		fileName := fmt.Sprintf("%v.%v",
			strings.Replace(name, " ", "_", -1)+"_"+fmt.Sprintf("%v", time.Now().Unix()),
			extension)
	
		if err := ctr.cloud.UploadFileToCloud(content, fileName, cf.ContractTypeFolderGCS); err != nil {
			return "", fmt.Errorf("upload template file to GCS got err: %v", err)
		}
	
		return fileName, nil
	}
	