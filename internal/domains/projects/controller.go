package projects

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
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
)

// ProjectController struct implements ProjectRepository
type ProjectController struct {
	cm.BaseController
	ProjRepo        rp.ProjectRepository
	UserRepo        rp.UserRepository
	UserProjectRepo rp.UserProjectRepository
	cloud           gc.StorageUtility
}

// NewProjectController import projRepo
func NewProjectController(logger echo.Logger, projRepo rp.ProjectRepository, userRepo rp.UserRepository, userProject rp.UserProjectRepository, cloud gc.StorageUtility) (ctr *ProjectController) {
	ctr = &ProjectController{cm.BaseController{}, projRepo, userRepo, userProject, cloud}
	ctr.Init(logger)
	return
}

// AddProject : add new Project to database
// Params : echo.Context
// Returns : return error
func (ctr *ProjectController) AddProject(c echo.Context) error {
	createProjectParams := new(param.CreateProjectParams)
	if err := c.Bind(createProjectParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	userProfile := c.Get("user_profile").(m.User)

	if _, err := valid.ValidateStruct(createProjectParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	project, err := ctr.ProjRepo.SaveProject(userProfile.OrganizationID, createProjectParams, ctr.UserProjectRepo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Create Project Failed",
		})
	}

	projectResponse := map[string]interface{}{
		"project_id":          project.ID,
		"project_name":        project.Name,
		"project_description": project.Description,
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Project Created Successful",
		Data:    projectResponse,
	})
}

// GetProjectList : get list of projects by Organization (by projectName keyword)
// Params : echo.Context
// Returns : return error
func (ctr *ProjectController) GetProjectList(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	projectListParams := new(param.ProjectListParams)

	if err := c.Bind(projectListParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(projectListParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	projects, totalRow, err := ctr.ProjRepo.GetProjects(userProfile.OrganizationID, projectListParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get project list failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if projectListParams.RowPerPage == 0 {
		projectListParams.CurrentPage = 1
		projectListParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": projectListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": projectListParams.RowPerPage,
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

	userList := make(map[int]string)

	for i := 0; i < len(users); i++ {
		userList[users[i].UserID] = users[i].FullName
	}

	listProjectResponse := []map[string]interface{}{}
	for i := 0; i < len(projects); i++ {
		itemDataResponse := map[string]interface{}{
			"project_id":          projects[i].ID,
			"project_name":        projects[i].Name,
			"project_description": projects[i].Description,
			"project_targets":     projects[i].Targets,
			"managed_by":          userList[projects[i].ManagedBy],
			"created_at":          projects[i].CreatedAt.Format(cf.FormatDateDisplay),
			"updated_at":          projects[i].UpdatedAt.Format(cf.FormatDateDisplay),
		}

		listProjectResponse = append(listProjectResponse, itemDataResponse)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination": pagination,
			"projects":   listProjectResponse,
			"users":      userList,
		},
	})
}

// GetProjectOfUser : get list of projects by Organization (by projectName keyword)
// Params : echo.Context
// Returns : return error
func (ctr *ProjectController) GetProjectOfUser(c echo.Context) error {
	userProfile := c.Get("user_profile").(m.User)
	projectListParams := new(param.ProjectListParams)

	if err := c.Bind(projectListParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(projectListParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	projects, totalRow, err := ctr.ProjRepo.GetProjectsOfUser(userProfile.UserProfile.UserID, userProfile.RoleID, userProfile.OrganizationID, projectListParams)

	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return c.JSON(http.StatusOK, cf.JsonResponse{
				Status:  cf.FailResponseCode,
				Message: "Get project list failed",
			})
		}

		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if projectListParams.RowPerPage == 0 {
		projectListParams.CurrentPage = 1
		projectListParams.RowPerPage = totalRow
	}

	pagination := map[string]interface{}{
		"current_page": projectListParams.CurrentPage,
		"total_row":    totalRow,
		"row_per_page": projectListParams.RowPerPage,
	}

	listProjectResponse := []map[string]interface{}{}
	for i := 0; i < len(projects); i++ {
		itemDataResponse := map[string]interface{}{
			"project_id":   projects[i].Id,
			"project_name": projects[i].Name,
			"joined_at":    projects[i].JoinedAt,
		}

		listProjectResponse = append(listProjectResponse, itemDataResponse)
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data: map[string]interface{}{
			"pagination": pagination,
			"projects":   listProjectResponse,
		},
	})
}

// UpdateProject : update project
// Params : echo.Context
// Returns : return error
func (ctr *ProjectController) UpdateProject(c echo.Context) error {
	updateProjectParams := new(param.UpdateProjectParams)
	if err := c.Bind(updateProjectParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(updateProjectParams); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	_, er := ctr.ProjRepo.GetProjectByID(updateProjectParams.ID)

	if er != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project not found",
			Data:    er,
		})
	}

	if err := ctr.ProjRepo.UpdateProject(updateProjectParams, ctr.UserProjectRepo); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Update project success",
	})
}

// GetProjectDetails : get project by ID
// Params : echo.Context
// Returns : return object
func (ctr *ProjectController) GetProjectDetails(c echo.Context) error {
	projectIDParam := new(param.ProjectIDParam)
	if err := c.Bind(projectIDParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(projectIDParam); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	project, err := ctr.ProjRepo.GetProjectByID(projectIDParam.ProjectID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	users, err := ctr.UserRepo.GetAllUserNameByOrgID(userProfile.OrganizationID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
			Data:    err,
		})
	}

	var usersId []int
	userList := make(map[int]string)
	for i := 0; i < len(users); i++ {
		userList[users[i].UserID] = users[i].FullName
		usersId = append(usersId, users[i].UserID)
	}

	userRecords, err := ctr.UserProjectRepo.SelectMembersInProject(userProfile.OrganizationID, projectIDParam.ProjectID)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	var usersIdJoinProject []int
	if len(userRecords) > 0 {
		for _, user := range userRecords {
			usersIdJoinProject = append(usersIdJoinProject, user.UserId)
		}
	}

	projectResponse := map[string]interface{}{
		"project_id":            project.ID,
		"name":                  project.Name,
		"description":           project.Description,
		"targets":               project.Targets,
		"managed_by":            project.ManagedBy,
		"users":                 userList,
		"users_id_join_project": usersIdJoinProject,
		"avatars":               ctr.SelectAssigneeAndAvatars(usersId),
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Success",
		Data:    projectResponse,
	})
}

// DeleteProject : delete project by id
// Params : echo.Context
// Returns : object
func (ctr *ProjectController) DeleteProject(c echo.Context) error {
	projectIDParam := new(param.ProjectIDParam)
	if err := c.Bind(projectIDParam); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(projectIDParam); err != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: err.Error(),
		})
	}

	_, er := ctr.ProjRepo.GetProjectByID(projectIDParam.ProjectID)

	if er != nil {
		return c.JSON(http.StatusOK, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Project not found",
			Data:    er,
		})
	}

	userProfile := c.Get("user_profile").(m.User)
	err := ctr.UserProjectRepo.DeleteByProjectId(userProfile.OrganizationID, projectIDParam.ProjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	err = ctr.ProjRepo.DeleteProject(projectIDParam.ProjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Deleted",
	})
}

func (ctr *ProjectController) SelectAssigneeAndAvatars(assignees []int) map[int][]byte {
	records := make(map[int][]byte)
	if len(assignees) == 0 {
		return records
	}

	userIdAndAvatars, err := ctr.UserRepo.SelectAvatarUsers(assignees)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		panic(err)
	}

	for _, user := range userIdAndAvatars {
		var base64Img []byte
		if user.Avatar != "" {
			base64Img, err = ctr.cloud.GetFileByFileName(user.Avatar, cf.AvatarFolderGCS)
			if err != nil {
				ctr.Logger.Error(err)
			}
		}

		records[user.UserId] = base64Img
	}

	return records
}
