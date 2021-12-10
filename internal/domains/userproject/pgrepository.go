package userproject

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"strconv"
)

type PgUserProjectRepository struct {
	cm.AppRepository
}

// NewPgUserProjectRepository : Init PgUserProjectRepository
func NewPgUserProjectRepository(logger echo.Logger) (repo *PgUserProjectRepository) {
	repo = &PgUserProjectRepository{}
	repo.Init(logger)
	return
}

// InsertUserProject : Insert user project
func (repo *PgUserProjectRepository) InsertUserProject(userID int, projectID int) error {
	userProject := m.UserProject{
		UserID:    userID,
		ProjectID: projectID,
	}

	err := repo.DB.Insert(&userProject)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// InsertUserProject : Insert user project
func (repo *PgUserProjectRepository) InsertUserProjectWithTx(tx *pg.Tx, userID int, projectID int) error {
	userProject := m.UserProject{
		UserID:    userID,
		ProjectID: projectID,
	}

	err := tx.Insert(&userProject)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// CountUserProject : Count user project
func (repo *PgUserProjectRepository) CountUserProject(organizationID int) ([]param.NumberPeopleProject, error) {
	var records []param.NumberPeopleProject
	err := repo.DB.Model(&m.UserProject{}).
		ColumnExpr("p.id AS project_id").
		ColumnExpr("p.name AS project_name").
		ColumnExpr("p.managed_by AS managed_by").
		ColumnExpr("COUNT(p.name) AS amount").
		Join("JOIN users AS u ON u.id = user_project.user_id").
		Join("JOIN projects AS p ON p.id = user_project.project_id").
		Where("u.organization_id = ?", organizationID).
		Group("p.name", "p.created_at").
		Group("p.name", "p.id", "p.managed_by", "p.created_at").
		Order("p.created_at DESC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

// SelectUserProject : Select user project
func (repo *PgUserProjectRepository) SelectUserProject(organizationId int, getUserProjectParams *param.GetUserProjectParams) ([]param.UserForEachProject, error) {
	var userForEachProject []param.UserForEachProject
	err := repo.DB.Model(&m.UserProject{}).
		Column("user_project.id", "user_project.user_id", "user_project.created_at").
		ColumnExpr("up.branch").
		Join("JOIN user_profiles AS up ON up.user_id = user_project.user_id").
		Join("JOIN users AS u ON u.id = user_project.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_project.project_id = ?", getUserProjectParams.ProjectID).
		Select(&userForEachProject)

	if err != nil {
		repo.Logger.Error(err)
	}

	return userForEachProject, err
}

// CountUserProject : Count user project
func (repo *PgUserProjectRepository) RemoveUserProject(organizationId int, Id int) error {
	_, err := repo.DB.Model(&m.UserProject{}).
		TableExpr("users AS u").
		Where("u.id = user_project.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_project.id = ?", Id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// CheckExistUserFromProject: Check user from project
func (repo *PgUserProjectRepository) CheckExistUserFromProject(organizationId int, userID int, projectID int) (int, error) {
	count, err := repo.DB.Model(&m.UserProject{}).
		Join("JOIN users AS u ON u.id = user_project.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_project.user_id = ?", userID).
		Where("user_project.project_id = ?", projectID).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

// CheckExistUserFromProject: Check user from project
func (repo *PgUserProjectRepository) SelectUserIdsManagedByManager(organizationId int, userIdOfManager int) ([]param.UserProjectInfoRecords, error) {
	var userIds []param.UserProjectInfoRecords
	err := repo.DB.Model(&m.UserProject{}).
		ColumnExpr("DISTINCT user_project.user_id").
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		Join("JOIN projects AS p ON p.id = user_project.project_id").
		Join("JOIN users AS u ON u.id = user_project.user_id").
		Join("JOIN user_profiles AS up ON up.user_id = user_project.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("p.managed_by = ?", userIdOfManager).
		Select(&userIds)

	if err != nil {
		repo.Logger.Error(err)
	}

	return userIds, err
}

func (repo *PgUserProjectRepository) DeleteByProjectId(organizationId int, projectId int) error {
	_, err := repo.DB.Model(&m.UserProject{}).
		TableExpr("users AS u").
		Where("u.organization_id = ?", organizationId).
		Where("user_project.project_id = ?", projectId).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgUserProjectRepository) SelectProjectsByUserId(organizationId int, userId int) ([]m.Project, error) {
	var projects []m.Project
	err := repo.DB.Model(&m.UserProject{}).
		ColumnExpr("p.id, p.name, p.description, p.targets, p.managed_by, p.created_at, p.updated_at").
		Join("JOIN projects AS p ON p.id = user_project.project_id").
		Join("JOIN users AS u ON u.id = user_project.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_project.user_id = ?", userId).
		Select(&projects)

	if err != nil {
		repo.Logger.Error(err)
	}

	return projects, err
}

func (repo *PgUserProjectRepository) SelectUserIdsJoinProjectsWithUserId(organizationId int, userId int) ([]int, error) {
	var records []int
	q := "SELECT DISTINCT user_id " +
		"FROM user_projects AS upj " +
		"JOIN users AS u ON u.id = upj.user_id " +
		"WHERE u.organization_id = " + strconv.Itoa(organizationId) + " " +
		"AND project_id IN (" +
		"SELECT upj2.project_id " +
		"FROM user_projects AS upj2 " +
		"WHERE upj2.user_id = " + strconv.Itoa(userId) + " " +
		"AND upj2.deleted_at IS NULL)"

	_, err := repo.DB.Query(&records, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserProjectRepository) SelectMembersInProject(organizationId int, projectId int) ([]param.UserProjectInfoRecords, error) {
	var users []param.UserProjectInfoRecords
	err := repo.DB.Model(&m.UserProject{}).
		Column("user_project.user_id").
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		Join("JOIN projects AS p ON p.id = user_project.project_id").
		Join("JOIN user_profiles AS up ON up.user_id = user_project.user_id").
		Where("p.organization_id = ?", organizationId).
		Where("user_project.project_id = ?", projectId).
		Select(&users)

	if err != nil {
		repo.Logger.Error(err)
	}

	return users, err
}
