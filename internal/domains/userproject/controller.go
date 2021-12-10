package userproject

import (
	"net/http"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// Controller : User project Controller
type Controller struct {
	cm.BaseController

	UserProjectRepo rp.UserProjectRepository
	UserRepo        rp.UserRepository
	ProjectRepo     rp.ProjectRepository
	BranchRepo      rp.BranchRepository
}

// NewUserProjectController : Init UserProject Controller
func NewUserProjectController(
	logger echo.Logger,
	userProjectRepo rp.UserProjectRepository,
	userRepo rp.UserRepository,
	projectRepo rp.ProjectRepository,
	branchRepo rp.BranchRepository,
) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, userProjectRepo, userRepo, projectRepo, branchRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateUserProject(c echo.Context) error {
	insertUserProjectParams := new(param.InsertUserProjectParams)
	if err := c.Bind(insertUserProjectParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(insertUserProjectParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	count, err := ctr.UserProjectRepo.CheckExistUserFromProject(userProfile.OrganizationID, insertUserProjectParams.UserID, insertUserProjectParams.ProjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusNotAcceptable, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "User already exist in the project",
		})
	}

	err = ctr.UserProjectRepo.InsertUserProject(insertUserProjectParams.UserID, insertUserProjectParams.ProjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create user project successfully.",
	})
}

// GetUserProject: Get users for each project
func (ctr *Controller) GetUserProject(c echo.Context) error {
	getUserProjectParams := new(param.GetUserProjectParams)
	if err := c.Bind(getUserProjectParams); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(getUserProjectParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.UserProjectRepo.SelectUserProject(userProfile.OrganizationID, getUserProjectParams)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "User Project is not exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	var datasets []map[string]interface{}
	for _, record := range records {
		data := map[string]interface{}{
			"id":          record.ID,
			"user_id":     record.UserID,
			"date_joined": record.CreatedAt.Format(cf.FormatDateDisplay),
			"branch":      record.Branch,
			"is_deleted":  false,
		}

		datasets = append(datasets, data)
	}

	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
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

	var userBranchList []map[string]int
	for _, user := range users {
		userBranch := map[string]int{
			"user_id": user.UserID,
			"branch":  user.Branch,
		}

		userBranchList = append(userBranchList, userBranch)
	}

	branchRecords, err := ctr.BranchRepo.SelectBranches(userProfile.OrganizationID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Empty record.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	branches := make(map[int]string)
	for _, record := range branchRecords {
		branches[record.Id] = record.Name
	}

	dataResponse := map[string]interface{}{
		"user_box":         userList,
		"branch_box":       branches,
		"user_branch_list": userBranchList,
		"user_projects":    datasets,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get user project successfully.",
		Data:    dataResponse,
	})
}

// RemoveUserProject: Remove user from project
func (ctr *Controller) RemoveUserProject(c echo.Context) error {
	removeUserFromProjectParam := new(param.RemoveUserFromProjectParam)
	if err := c.Bind(removeUserFromProjectParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(removeUserFromProjectParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err = ctr.UserProjectRepo.RemoveUserProject(userProfile.OrganizationID, removeUserFromProjectParam.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Remove user from project successfully.",
	})
}

// GetUserIdManagedByManager: Get user ids managed by manager
func (ctr *Controller) GetUserIdsManagedByManager(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	records, err := ctr.UserProjectRepo.SelectUserIdsManagedByManager(userProfile.OrganizationID, userProfile.UserProfile.UserID)
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "There are no users to manage.",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var userIds []int
	for _, record := range records {
		userIds = append(userIds, record.UserId)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get user ids successfully.",
		Data:    userIds,
	})
}

func (ctr *Controller) GetProjectsUserJoin(c echo.Context) error {
	projectUserJoinParam := new(param.ProjectUserJoinParam)
	if err := c.Bind(projectUserJoinParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
		})
	}

	_, err := valid.ValidateStruct(projectUserJoinParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	projects, err := ctr.UserProjectRepo.SelectProjectsByUserId(
		userProfile.OrganizationID,
		projectUserJoinParam.UserId,
	)

	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var dataResponse []map[string]interface{}
	for _, project := range projects {
		data := map[string]interface{}{
			"project_id":          project.ID,
			"project_name":        project.Name,
			"project_description": project.Description,
			"project_targets":     project.Targets,
			"managed_by":          project.ManagedBy,
			"created_at":          project.CreatedAt.Format(cf.FormatDateDisplay),
			"updated_at":          project.UpdatedAt.Format(cf.FormatDateDisplay),
		}

		dataResponse = append(dataResponse, data)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get projects successfully",
		Data:    dataResponse,
	})
}

func (ctr *Controller) ProjectStatistic(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	numberPeopleProject, err := ctr.UserProjectRepo.CountUserProject(userProfile.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System error",
		})
	}

	dataResponse := map[string]interface{}{
		"number_people_project": numberPeopleProject,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Project statistic successful",
		Data:    dataResponse,
	})
}
