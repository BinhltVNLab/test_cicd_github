package projects

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// PgProjRepository : struct for project repository
type PgProjRepository struct {
	cm.AppRepository
}

// NewPgProjRepository : struct for initing new ProjectRepository
func NewPgProjRepository(logger echo.Logger) (repo *PgProjRepository) {
	repo = &PgProjRepository{}
	repo.Init(logger)
	return
}

// SaveProject : insert data to project
// Params : orgID, param.CreateProjectParams
// Returns : return object of record that 've just been inserted
func (repo *PgProjRepository) SaveProject(
	orgID int,
	createProjectParams *param.CreateProjectParams,
	userProjectRepo rp.UserProjectRepository,
) (m.Project, error) {
	project := m.Project{}
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		project, transErr = repo.InsertProjectWithTx(
			tx,
			createProjectParams.ProjectName,
			createProjectParams.ProjectDescription,
			createProjectParams.ManagedBy,
			orgID,
		)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}

		transErr = userProjectRepo.InsertUserProjectWithTx(tx, createProjectParams.ManagedBy, project.ID)
		if transErr != nil {
			repo.Logger.Error(transErr)
			return transErr
		}
		return transErr
	})

	return project, err
}

// InsertProjectWithTx : insert data to projects
// Params : pg.Tx, projectName, projectDescription, organizationID
// Returns : return project object , error
func (repo *PgProjRepository) InsertProjectWithTx(tx *pg.Tx, projectName string, projectDescription string, managedBy int, organizationID int) (m.Project, error) {
	project := m.Project{
		Name:           projectName,
		OrganizationID: organizationID,
		Description:    projectDescription,
		ManagedBy:      managedBy,
	}
	err := tx.Insert(&project)
	return project, err
}

// GetProjectsByOrganizationID : get projects list by org ID (by projectName keyword)
// Params : orgID
// Returns : []project, totalRow, err
func (repo *PgProjRepository) GetProjects(orgID int, projectListParams *param.ProjectListParams) ([]m.Project, int, error) {
	projects := []m.Project{}
	queryObj := repo.DB.Model(&projects)
	queryObj.Where("organization_id = ?", orgID)
	if projectListParams.Keyword != "" {
		queryObj.Where("LOWER(name) LIKE LOWER(?)", "%"+projectListParams.Keyword+"%")
	}
	queryObj.Offset((projectListParams.CurrentPage - 1) * projectListParams.RowPerPage)
	queryObj.Order("created_at DESC")
	queryObj.Limit(projectListParams.RowPerPage)
	totalRow, err := queryObj.SelectAndCount()
	return projects, totalRow, err
}

func (repo *PgProjRepository) GetProjectsOfUser(userID int, roleID int, orgID int, projectListParams *param.ProjectListParams) ([]param.ProjectUserJoinRecords, int, error) {
	var projects = []m.Project{}
	var projectUserJoinRecords []param.ProjectUserJoinRecords

	queryObj := repo.DB.Model(&projects)
	queryObj.ColumnExpr("prj.id")
	queryObj.ColumnExpr("prj.name")

	if roleID != cf.GeneralManagerRoleID {
		queryObj.ColumnExpr("date(uprj.created_at) as joined_at")
		queryObj.Join("JOIN user_projects AS uprj ON prj.id = uprj.project_id")
		queryObj.Where("uprj.deleted_at IS NULL")
	} else {
		queryObj.ColumnExpr("date(prj.created_at) as joined_at")
	}
	queryObj.Where("prj.organization_id = ?", orgID)
	
	if roleID != cf.GeneralManagerRoleID {
		queryObj.Where("uprj.user_id = ?", userID)
	}
	if projectListParams.Keyword != "" {
		queryObj.Where("LOWER(prj.name) LIKE LOWER(?)", "%"+projectListParams.Keyword+"%")
	}
	queryObj.Group("prj.id", "joined_at")
	queryObj.Order("joined_at ASC")
	if (projectListParams.CurrentPage != 0 && projectListParams.RowPerPage != 0) {
		queryObj.Offset((projectListParams.CurrentPage - 1) * projectListParams.RowPerPage)
		queryObj.Limit(projectListParams.RowPerPage)
	}
	totalRow, err := queryObj.SelectAndCount(&projectUserJoinRecords)

	if err != nil {
		repo.Logger.Error(err)
	}

	return projectUserJoinRecords, totalRow, err
}

// UpdateProject : update project
// Params : ProjectID, ProjectName, ProjectDescription, ProjectTargets[]
// Returns : error
func (repo *PgProjRepository) UpdateProject(projParams *param.UpdateProjectParams, userProjectRepo rp.UserProjectRepository) error {
	currentProject, err := repo.GetProjectByID(projParams.ID)
	if err != nil {
		repo.Logger.Error()
		return err
	}

	project := &m.Project{
		Name:        projParams.Name,
		Description: projParams.Description,
		Targets:     projParams.Targets,
		ManagedBy:   projParams.ManagedBy,
	}

	users, transErr := userProjectRepo.SelectMembersInProject(currentProject.OrganizationID, projParams.ID)
	if transErr != nil && transErr.Error() != pg.ErrNoRows.Error() {
		repo.Logger.Error()
		return transErr
	}

	var usersId []int
	if len(users) > 0 {
		for _, user := range users {
			usersId = append(usersId, user.UserId)
		}
	}

	if projParams.ManagedBy != currentProject.ManagedBy && !utils.FindIntInSlice(usersId, projParams.ManagedBy) {
		err = repo.DB.RunInTransaction(func(tx *pg.Tx) error {
			var transErr error
			if _, transErr = tx.Model(project).
				Column("name", "description", "targets", "managed_by", "updated_at").
				Where("id = ?", projParams.ID).
				Update(); transErr != nil {
				repo.Logger.Error()
				return transErr
			}

			if transErr = userProjectRepo.InsertUserProjectWithTx(tx, projParams.ManagedBy, projParams.ID); transErr != nil {
				repo.Logger.Error()
				return transErr
			}

			return transErr
		})
	} else {
		_, err = repo.DB.Model(project).
			Column("name", "description", "targets", "managed_by", "updated_at").
			Where("id = ?", projParams.ID).
			Update()
	}

	return err
}

// GetProjectByID : get project by ID
// Params : projectID
// Returns : object, error
func (repo *PgProjRepository) GetProjectByID(projectID int) (m.Project, error) {
	project := m.Project{}
	err := repo.DB.Model(&project).
		Column("prj.*").
		Where("prj.id = ?", projectID).
		First()

	return project, err
}

// DeleteProject : delete project by ID
// Params : projectID
// Returns : error
func (repo *PgProjRepository) DeleteProject(projectID int) error {
	project := m.Project{}
	_, err := repo.DB.Model(&project).
		Where("id = ?", projectID).
		Delete()

	return err
}

// CountProject : count total project
func (repo *PgProjRepository) CountProject(organizationID int) (int, error) {
	var total int
	err := repo.DB.Model(&m.Project{}).
		ColumnExpr("COUNT(id) AS total").
		Where("organization_id = ?", organizationID).
		Select(&total)

	if err != nil {
		repo.Logger.Error(err)
	}

	return total, err
}

func (repo *PgProjRepository) SelectProjectsByOrganizationId(organizationId int) ([]m.Project, error) {
	var projects []m.Project
	err := repo.DB.Model(&projects).
		Column("id", "name").
		Where("organization_id = ?", organizationId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return projects, err
}

func (repo *PgProjRepository) SelectProjectManagers(organizationId int) ([]int, error) {
	var records []int
	err := repo.DB.Model(&m.Project{}).
		Column("managed_by").
		Where("organization_id = ?", organizationId).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgProjRepository) CountProjectById(projectId int) (int, error) {
	count, err := repo.DB.Model(&m.Project{}).
		Where("id = ?", projectId).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

