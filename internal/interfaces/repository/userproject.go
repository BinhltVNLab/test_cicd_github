package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// UserProjectRepository interface
type UserProjectRepository interface {
	InsertUserProject(userID int, projectID int) error
	InsertUserProjectWithTx(tx *pg.Tx, userID int, projectID int) error
	CountUserProject(organizationID int) ([]param.NumberPeopleProject, error)
	SelectUserProject(organizationId int, getUserProjectParams *param.GetUserProjectParams) ([]param.UserForEachProject, error)
	RemoveUserProject(organizationId int, Id int) error
	CheckExistUserFromProject(organizationId int, userID int, projectID int) (int, error)
	SelectUserIdsManagedByManager(organizationId int, userIdOfManager int) ([]param.UserProjectInfoRecords, error)
	DeleteByProjectId(organizationId int, projectId int) error
	SelectProjectsByUserId(organizationId int, userId int) ([]m.Project, error)
	SelectUserIdsJoinProjectsWithUserId(organizationId int, userId int) ([]int, error)
	SelectMembersInProject(organizationId int, projectId int) ([]param.UserProjectInfoRecords, error)
}
