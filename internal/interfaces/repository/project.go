package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// ProjectRepository interface
type ProjectRepository interface {
	SaveProject(orgID int, createProjectParams *param.CreateProjectParams, userProjectRepo UserProjectRepository) (m.Project, error)
	InsertProjectWithTx(tx *pg.Tx, projectName string, projectDescription string, managedBy int, organizationID int) (m.Project, error)
	GetProjects(orgID int, projectListParams *param.ProjectListParams) ([]m.Project, int, error)
	GetProjectsOfUser(userID int, roleID int, orgID int, projectListParams *param.ProjectListParams) ([]param.ProjectUserJoinRecords, int, error)
	UpdateProject(projectParams *param.UpdateProjectParams, userProjectRepo UserProjectRepository) error
	GetProjectByID(projectID int) (m.Project, error)
	DeleteProject(projectID int) error
	CountProject(organizationID int) (int, error)
	SelectProjectsByOrganizationId(organizationId int) ([]m.Project, error)
	SelectProjectManagers(organizationId int) ([]int, error)
	CountProjectById(projectId int) (int, error)
}
