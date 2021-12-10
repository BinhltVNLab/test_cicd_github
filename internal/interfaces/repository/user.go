package repository

import (
	"time"

	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// UserRepository interface
type UserRepository interface {
	GetUserProfile(id int) (m.User, error)
	GetUserProfileExpand(id int) (m.UserProfileExpand, error)
	GetUser(id int) (m.User, error)
	GetUserByEmailOrganizationID(email string, orgID int) (m.User, error)
	GetUserByResetCode(code string) (m.User, error)
	GetUserByChangeEmailCode(code string) (m.User, error)
	GetLoginUserID(email string, password string, organizationID int) (int, error)
	UpdateLastLogin(userID int) error
	UpdateResetPassword(resetPasswordParams *param.ResetPasswordParams) error
	InsertUserProfileWithTx(tx *pg.Tx, userID int, firstname string, lastname string) error
	InsertUserWithTx(tx *pg.Tx, registerUser param.RegisterUserParams) (m.User, error)
	SetForgotPassParams(setForgotPassParams param.SetForgotPassParams) error
	CheckCurrentPassword(userID int, password string) (bool, error)
	UpdatePasswordByID(userID int, password string) error
	SetNewUpdateEmail(setUpdateEmailParams param.SetUpdateEmailParams, userID int) error
	UpdateEmail(userObj m.User) error
	GetUserProfileList(orgID int, userProfileListParams *param.UserProfileListParams) ([]m.UserExt, int, error)
	SaveProfile(basicProfileObj *param.EditProfileParams) error
	GetAllUserNameByOrgID(orgID int) ([]param.AllUserName, error)
	GetAllUserNameAndCountByOrgID(allUserNameAndCountParams *param.AllUserNameAndCountParams) ([]param.AllUserName, int, error)
	InsertIntoUserRankLog(userID int, rank int) error
	GetUserRankLogByUserID(userID int) ([]m.UserRankLog, error)
	CountPeopleBranch(organizationID int) ([]param.NumberPeopleEachBranch, error)
	CountPeopleJobTitle(organizationID int) ([]param.NumberPeopleJobTitle, error)
	CountUser(organizationID int) (int, error)
	CountJapaneseLevel(organizationID int) ([]param.NumberPeopleJpLanguageCert, error)
	UpdateIntFieldToNull(organizationID int, field string, valueWhere int) error
	SelectUserIdByEmployeeId(organizationId int, employeeId string) (int, error)
	UpdateProfileWithTx(updateProfileParams *[]param.UpdateProfileParams) error
	UpdateLanguageSetting(organizationId int, languageSettingParams *param.LanguageSettingParams) error
	SelectEmployeeIdByOrganizationId(organizationId int) ([]param.EmployeeIdAndFullName, error)
	SelectEmailOfGMAndPM(organizationId int) ([]param.EmailOfGMAndPMRecords, error)
	SelectUsersByJpLevel(
		organizationID int,
		jpLevelStatisticDetailParams *param.JpLevelStatisticDetailParams,
	) ([]param.FullStatisticDetail, int, error)
	SelectIdsOfGMAndManager(organizationId int) ([]int, error)
	SelectFullNameUser(userId int) (string, error)
	SelectAvatarUsers(usersId []int) ([]param.UserIdAndAvatarRecord, error)
	SelectEmailByUserIds(userIds []int) ([]string, error)
	SelectIdsOfGM(organizationId int) ([]int, error)
	GetUserProfileListByOrgId(organizationId int) ([]m.UserExt, error)
	UpdateContractExpireDate(userID int, contractExpirationDate time.Time) error
}
