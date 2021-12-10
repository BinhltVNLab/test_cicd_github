package users

import (
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type PgUserRepository struct {
	cm.AppRepository
}

func NewPgUserRepository(logger echo.Logger) (repo *PgUserRepository) {
	repo = &PgUserRepository{}
	repo.Init(logger)
	return
}

// GetUser : get user
// Params  : user id - POST
// Returns : user (Object)
func (repo *PgUserRepository) GetUser(id int) (m.User, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Where("id = ?", id).
		First()

	return user, err
}

// GetUserByEmailOrganizationID : get user by email and org ID
// Params  :
// Returns : user (Object)
func (repo *PgUserRepository) GetUserByEmailOrganizationID(email string, orgID int) (m.User, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Where("email = ?", email).
		Where("organization_id = ?", orgID).
		First()

	return user, err
}

// GetUserByResetCode : get user by reset password code
// Params  :
// Returns : user (Object)
func (repo *PgUserRepository) GetUserByResetCode(code string) (m.User, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Where("reset_password_code = ?", code).
		First()

	return user, err
}

// GetUserByChangeEmailCode : get user by change email code
// Params  :
// Returns : user (Object)
func (repo *PgUserRepository) GetUserByChangeEmailCode(code string) (m.User, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Where("update_email_code = ?", code).
		First()

	return user, err
}

// GetUserProfile : get user profile include role, organization name
// Params  : user id - POST
// Returns : user profile (Object)
func (repo *PgUserRepository) GetUserProfile(id int) (m.User, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Column("usr.*").
		Where("usr.id = ?", id).
		Where("usr.deleted_at is null").
		Relation("UserProfile").
		Relation("Role").
		Relation("Organization").
		First()

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return user, err
}

// GetUserProfile : get user profile include role, organization name
// Params  : user id - POST
// Returns : user profile (Object)
func (repo *PgUserRepository) GetUserProfileExpand(id int) (m.UserProfileExpand, error) {
	user := m.UserProfileExpand{}
	err := repo.DB.Model(&user).
		ColumnExpr("usr.id, usr.organization_id, usr.email, usr.role_id, rol.name as role_name").
		ColumnExpr("pro.avatar, pro.first_name, pro.last_name, pro.birthday, pro.rank, pro.job_title").
		ColumnExpr("pro.phone_number, pro.company_joined_date, pro.skill, pro.language, pro.education").
		ColumnExpr("pro.department, pro.award, pro.experience, pro.date_severance, pro.reasons_severance").
		ColumnExpr("pro.introduce, pro.branch, pro.employee_id").
		ColumnExpr("pro.current_address, pro.tax_code, pro.status").
		ColumnExpr("pro.identity_card, pro.occupation, pro.date_of_identity_card").
		ColumnExpr("pro.country, pro.work_place, pro.permanent_residence, pro.job_position, pro.address_of_emergency, pro.license_plates, pro.account_number_vcb, pro.book_number_bhxh").
		ColumnExpr("pro.place_of_birth, pro.gender, pro.marital_status").
		ColumnExpr("pro.user_birth_place, pro.id_type, pro.place_of_issue, pro.country, pro.nation, pro.religion, pro.name_of_emergency, pro.relationships_of_emergency").
		Join("JOIN user_profiles AS pro ON pro.user_id = usr.id").
		Join("JOIN organizations AS org ON org.id = usr.organization_id").
		Join("JOIN user_roles AS rol ON rol.id = usr.role_id").
		Where("usr.id = ?", id).
		First()

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return user, err
}

// GetLoginUserID : get ID User login
// Params  : email - POST, password - POST, organizationID - POST
// Returns : user id (int)
func (repo *PgUserRepository) GetLoginUserID(email string, password string, organizationID int) (int, error) {
	user := m.User{}
	err := repo.DB.Model(&user).
		Column("id").
		Where("email = ?", email).
		Where("password = ?", password).
		Where("organization_id = ?", organizationID).
		Where("deleted_at is null").
		Select()

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return user.ID, err
}

func (repo *PgUserRepository) UpdateLastLogin(userID int) error {
	_, err := repo.DB.Model(&m.User{LastLoginTime: utils.TimeNowUTC()}).
		Column("last_login_time", "updated_at").
		Where("id = ?", userID).
		Update()

	return err
}

// InsertUserProfileWithTx : insert data to user_profiles
// Params            : userID, firstname, lastname, currentTime - input data to insert
// Returns           : error
func (repo *PgUserRepository) InsertUserProfileWithTx(tx *pg.Tx, userID int, firstname string, lastname string) error {
	userProfile := m.UserProfile{
		UserID:    userID,
		FirstName: firstname,
		LastName:  lastname,
	}
	err := tx.Insert(&userProfile)

	return err
}

// InsertUserWithTx : insert data to users
// Params     : orgID, organizationTag, email, password - input data to insert
// Returns    : return users object , error
func (repo *PgUserRepository) InsertUserWithTx(tx *pg.Tx, registerUser param.RegisterUserParams) (m.User, error) {
	user := m.User{
		OrganizationID: registerUser.OrganizationID,
		Email:          registerUser.Email,
		Password:       utils.GetSHA256Hash(registerUser.Password),
		RoleID:         registerUser.RoleID,
		GoogleID:       registerUser.GoogleID,
		LastLoginTime:  utils.TimeNowUTC(),
		LanguageId:     registerUser.LanguageId,
	}
	err := tx.Insert(&user)

	return user, err
}

// SetForgotPassParams : set info user forgot password, code forgot, code code expired
// Params         : object include userid, code reset password, time expired
// Returns        : error
func (repo *PgUserRepository) SetForgotPassParams(setForgotPassParams param.SetForgotPassParams) error {
	user := &m.User{
		ResetPasswordCode: setForgotPassParams.ResetPasswordCode,
		CodeExpiredAt:     setForgotPassParams.CodeExpiredAt,
	}
	_, err := repo.DB.Model(user).
		Column("reset_password_code", "code_expired_at", "updated_at").
		Where("id = ?", setForgotPassParams.UserID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// UpdateResetPassword :
// Params            :
// Returns           :
func (repo *PgUserRepository) UpdateResetPassword(resetPasswordParams *param.ResetPasswordParams) error {
	user := &m.User{
		ResetPasswordCode: "",
		CodeExpiredAt:     utils.TimeNowUTC(),
		Password:          utils.GetSHA256Hash(resetPasswordParams.Password),
	}
	_, err := repo.DB.Model(user).
		Column("reset_password_code", "code_expired_at", "password", "updated_at").
		Where("reset_password_code = ?", resetPasswordParams.ResetPasswordCode).
		Update()

	return err
}

// SetNewUpdateEmail : insert new temporatory email wait for confirm
// Params            : SetUpdateEmailParams object - userID
// Returns           : update result
func (repo *PgUserRepository) SetNewUpdateEmail(setUpdateEmailParams param.SetUpdateEmailParams, userID int) error {
	user := &m.User{
		EmailForUpdate:           setUpdateEmailParams.EmailForUpdate,
		UpdateEmailCode:          setUpdateEmailParams.UpdateEmailCode,
		UpdateEmailCodeExpiredAt: setUpdateEmailParams.UpdateEmailCodeCodeExpiredAt,
	}
	_, err := repo.DB.Model(user).
		Column("email_for_update", "update_email_code", "update_email_code_expired_at", "updated_at").
		Where("id = ?", userID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// UpdateEmail : save temporatory email to login email
// Params      : User object
// Returns     : update result
func (repo *PgUserRepository) UpdateEmail(userObj m.User) error {
	user := &m.User{
		Email:                    userObj.EmailForUpdate,
		EmailForUpdate:           "",
		UpdateEmailCode:          "",
		UpdateEmailCodeExpiredAt: time.Time{},
		GoogleID:                 "",
	}

	_, err := repo.DB.Model(user).
		Column("email", "email_for_update", "update_email_code",
			"update_email_code_expired_at", "google_id", "updated_at").
		Where("id = ?", userObj.ID).
		Update()

	return err
}

// CheckCurrentPassword : check user input right current password before change new password
// Params            : current password
// Returns           : boolean, error
func (repo *PgUserRepository) CheckCurrentPassword(userID int, password string) (bool, error) {
	checkCurrentPassword := false

	user := m.User{}
	count, err := repo.DB.Model(&user).
		Where("id = ?", userID).
		Where("password = ?", utils.GetSHA256Hash(password)).
		Count()

	if count == 1 {
		checkCurrentPassword = true
	}

	return checkCurrentPassword, err
}

// UpdatePasswordByID : update password by ID
// Params            : user id, new password
// Returns           : error
func (repo *PgUserRepository) UpdatePasswordByID(userID int, password string) error {
	_, err := repo.DB.Model(&m.User{Password: utils.GetSHA256Hash(password)}).
		Column("password", "updated_at").
		Where("id = ?", userID).
		Update()

	return err
}

// GetUserProfileList : get user profile list
// Params             : orgID - organization ID , userProfileListParams - search param
// Returns            : UserProfileList object , total row , error
func (repo *PgUserRepository) GetUserProfileList(orgID int, userProfileListParams *param.UserProfileListParams) ([]m.UserExt, int, error) {
	var user []m.UserExt

	queryObj := repo.DB.Model(&user).
		Relation("Role").
		Relation("UserProfile").
		Relation("UserProfile.UserBranch").
		Where("usr.organization_id = ?", orgID)

	if userProfileListParams.Name != "" {
		profileName := "%" + userProfileListParams.Name + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(user_profile.first_name)) || ' ' || vietnamese_unaccent(LOWER(user_profile.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			profileName)
	}

	if userProfileListParams.Email != "" {
		queryObj.Where("LOWER(usr.email) LIKE LOWER(?)", "%"+userProfileListParams.Email+"%")
	}

	if userProfileListParams.PhoneNumber != "" {
		queryObj.Where("user_profile.phone_number LIKE ?", "%"+userProfileListParams.PhoneNumber+"%")
	}

	if !userProfileListParams.DateFrom.IsZero() {
		queryObj.Where("DATE(user_profile.company_joined_date) >= DATE(?)", userProfileListParams.DateFrom)
	}

	if !userProfileListParams.DateTo.IsZero() {
		queryObj.Where("DATE(user_profile.company_joined_date) <= DATE(?)", userProfileListParams.DateTo)
	}

	if userProfileListParams.Rank != 0 {
		queryObj.Where("user_profile.rank = ?", userProfileListParams.Rank)
	}

	if userProfileListParams.Branch != 0 {
		queryObj.Where("user_profile.branch = ?", userProfileListParams.Branch)
	}

	if userProfileListParams.Status != 0 {
		queryObj.Where("user_profile.status = ?", userProfileListParams.Status)
	}
	if userProfileListParams.JobTitle != 0 {
		queryObj.Where("user_profile.job_title = ?", userProfileListParams.JobTitle)
	}
	queryObj.Offset((userProfileListParams.CurrentPage - 1) * userProfileListParams.RowPerPage)
	queryObj.Order("user_profile.created_at DESC")
	queryObj.Limit(userProfileListParams.RowPerPage)
	totalRow, err := queryObj.SelectAndCount()

	if err != nil {
		repo.Logger.Error(err)
	}

	return user, totalRow, err
}

// SaveProfile : save basic profile
// Params            : user id, new password
// Returns           : error
func (repo *PgUserRepository) SaveProfile(basicProfileObj *param.EditProfileParams) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error = nil

		// update table profile
		birthdayDate, _ := time.Parse(cf.FormatDateDatabase, basicProfileObj.Birthday)
		companyJoinedDate, _ := time.Parse(cf.FormatDateDatabase, basicProfileObj.CompanyJoinedDate)
		issueDate, _ := time.Parse(cf.FormatDateDatabase, basicProfileObj.IssueDate)

		userProfile := m.UserProfile{}

		queryObj := tx.Model(&userProfile)

		// update basic profile
		if basicProfileObj.FlagEditBasicProfile {
			queryObj.Set("first_name = ?", basicProfileObj.FirstName).
				Set("last_name = ?", basicProfileObj.LastName).
				Set("birthday = ?", birthdayDate).
				Set("phone_number = ?", basicProfileObj.PhoneNumber).
				Set("job_title = ?", basicProfileObj.JobTitle).
				Set("company_joined_date = ?", companyJoinedDate).
				Set("branch = ?", basicProfileObj.Branch).
				Set("employee_id = ?", basicProfileObj.EmployeeId).
				Set("date_of_identity_card = ?", issueDate).
				Set("permanent_residence = ?", basicProfileObj.PermanentResidence).
				Set("current_address = ?", basicProfileObj.CurrentAddress).
				Set("identity_card = ?", basicProfileObj.IdentityCard).
				Set("tax_code = ?", basicProfileObj.TaxCode).
				Set("id_type = ?", basicProfileObj.IDType).
				Set("job_position = ?", basicProfileObj.JobPosition).
				Set("place_of_birth = ?", basicProfileObj.PlaceOfBirth).
				Set("gender = ?", basicProfileObj.Gender).
				Set("marital_status = ?", basicProfileObj.MaritalStatus).
				// Set("job_date = ?", basicProfileObj.JobDate).
				Set("book_number_bhxh = ?", basicProfileObj.BookNumberBhxh).
				Set("account_number_vcb = ?", basicProfileObj.AccountNumberVcb).
				Set("nation = ?", basicProfileObj.Nation).
				Set("religion = ?", basicProfileObj.Religion).
				Set("name_of_emergency = ?", basicProfileObj.NameOfEmergency).
				Set("relationships_of_emergency = ?", basicProfileObj.RelationshipsOfEmergency).
				Set("address_of_emergency = ?", basicProfileObj.AddressOfEmergency).
				Set("license_plates = ?", basicProfileObj.LicensePlates).
				Set("user_birth_place = ?", basicProfileObj.UserBirthPlace).
				Set("country = ?", basicProfileObj.Country).
				Set("place_of_issue = ?", basicProfileObj.PlaceOfIssue).
				Set("work_place = ?", basicProfileObj.WorkPlace)

			if basicProfileObj.DateSeverance != "" {
				t, _ := time.Parse(cf.FormatDateDatabase, basicProfileObj.DateSeverance)
				queryObj.Set("date_severance = ?", t)
			}

			if basicProfileObj.Department != nil {
				queryObj.Set("department = ?", *basicProfileObj.Department)
			}

			if basicProfileObj.ReasonsSeverance != nil {
				queryObj.Set("reasons_severance = ?", *basicProfileObj.ReasonsSeverance)
			}

			if basicProfileObj.Status != nil && *basicProfileObj.Status >= 0 {
				queryObj.Set("status = ?", *basicProfileObj.Status)
			}

			if basicProfileObj.FlagEditAvatar {
				queryObj.Set("avatar = ?", basicProfileObj.Avatar)
			}

			if basicProfileObj.EditorRoleID == cf.GeneralManagerRoleID {
				queryObj.Set("rank = ?", basicProfileObj.Rank)
			}

			// update role in table users
			if basicProfileObj.EditorRoleID == cf.GeneralManagerRoleID {
				_, errTx = tx.Model(&m.User{RoleID: basicProfileObj.RoleID}).
					Column("role_id", "updated_at").
					Where(`id = ?`, basicProfileObj.UserID).
					Update()

				if errTx != nil {
					repo.Logger.Error(errTx)
					return errTx
				}
			}
		}

		// update skill
		if basicProfileObj.FlagEditSkill {
			queryObj.Set("skill = ?", basicProfileObj.Skill)
		}

		if basicProfileObj.FlagEditLanguage {
			queryObj.Set("language = ?", basicProfileObj.Language)
		}

		if basicProfileObj.FlagEditEducation {
			queryObj.Set("education = ?", basicProfileObj.Education)
		}

		if basicProfileObj.FlagEditAward {
			queryObj.Set("award = ?", basicProfileObj.Award)
		}

		if basicProfileObj.FlagEditExperience {
			queryObj.Set("experience = ?", basicProfileObj.Experience)
		}

		if basicProfileObj.FlagEditIntroduce {
			queryObj.Set("introduce = ?", basicProfileObj.Introduce)
		}

		_, errTx = queryObj.Set("updated_at = ?", utils.TimeNowUTC()).
			Where("user_id = ?", basicProfileObj.UserID).
			Update()

		if errTx != nil {
			repo.Logger.Error(errTx)
			return errTx
		}

		return errTx
	})

	return err
}

// GetAllUserNameByOrgID : Get all user of organization
func (repo *PgUserRepository) GetAllUserNameByOrgID(orgID int) ([]param.AllUserName, error) {
	var userProfiles []m.UserProfile
	var records []param.AllUserName

	err := repo.DB.Model(&userProfiles).
		Column("user_profile.user_id", "user_profile.branch", "user_profile.company_joined_date",
		"user_profile.birthday", "user_profile.contract_expiration_date").
		ColumnExpr("user_profile.first_name || ' ' || user_profile.last_name full_name").
		Join("JOIN users AS u on u.id = user_profile.user_id").
		Where("user_profile.status != 3").
		Where("u.organization_id = ?", orgID).
		Select(&records)

	return records, err
}

// GetAllUserNameAndCountByOrgID : Get all user and count of organization
func (repo *PgUserRepository) GetAllUserNameAndCountByOrgID(
	allUserNameAndCountParams *param.AllUserNameAndCountParams,
) ([]param.AllUserName, int, error) {
	var userProfiles []m.UserProfile
	var records []param.AllUserName

	queryObj := repo.DB.Model(&userProfiles)
	queryObj.Column("user_profile.user_id")
	queryObj.ColumnExpr("user_profile.first_name || ' ' || user_profile.last_name full_name")
	queryObj.Column("user_profile.avatar")
	queryObj.Column("user_profile.branch")
	queryObj.Column("u.email")
	queryObj.Join("JOIN users AS u on u.id = user_profile.user_id")
	queryObj.Where("u.organization_id = ?", allUserNameAndCountParams.OrgID)
	queryObj.Where("user_profile.status != 3")

	if allUserNameAndCountParams.UserName != "" {
		userName := "%" + allUserNameAndCountParams.UserName + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(user_profile.first_name)) || ' ' || vietnamese_unaccent(LOWER(user_profile.last_name)) "+
			"LIKE vietnamese_unaccent(LOWER(?0))",
			userName)
	}

	if allUserNameAndCountParams.Branch != 0 {
		queryObj.Where("user_profile.branch = ?", allUserNameAndCountParams.Branch)
	}

	queryObj.Group("user_profile.user_id", "full_name", "user_profile.avatar", "u.email", "user_profile.branch")
	queryObj.Offset((allUserNameAndCountParams.CurrentPage - 1) * allUserNameAndCountParams.RowPerPage)
	queryObj.Limit(allUserNameAndCountParams.RowPerPage)
	queryObj.Order("user_profile.user_id ASC")
	totalRow, err := queryObj.SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgUserRepository) InsertIntoUserRankLog(userID int, rank int) error {
	record := m.UserRankLog{
		UserID: userID,
		Rank:   rank,
	}

	err := repo.DB.Insert(&record)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// GetUserRankLogByUserID : get rank log by user_id
func (repo *PgUserRepository) GetUserRankLogByUserID(userID int) ([]m.UserRankLog, error) {
	var records []m.UserRankLog
	err := repo.DB.Model(&records).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Select()
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

// CountPeopleBranch : Count the number of people each branch
func (repo *PgUserRepository) CountPeopleBranch(organizationID int) ([]param.NumberPeopleEachBranch, error) {
	var records []param.NumberPeopleEachBranch
	err := repo.DB.Model(&m.UserProfile{}).
		ColumnExpr("b.name AS branch").
		ColumnExpr("COUNT(b.name) AS amount").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Join("JOIN branches AS b ON b.id = user_profile.branch").
		Where("b.deleted_at IS NULL").
		Where("u.organization_id = ?", organizationID).
		Where("user_profile.branch > 0").
		Where("user_profile.status = 1 OR user_profile.status = 2").
		Group("b.name", "b.priority").
		Order("b.priority ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

// CountPeopleJobTitle : Count the number of people job title
func (repo *PgUserRepository) CountPeopleJobTitle(organizationID int) ([]param.NumberPeopleJobTitle, error) {
	var records []param.NumberPeopleJobTitle
	err := repo.DB.Model(&m.UserProfile{}).
		ColumnExpr("jt.name AS job_title").
		ColumnExpr("COUNT(jt.name) AS amount").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Join("JOIN job_titles AS jt ON jt.id = user_profile.job_title").
		Where("jt.deleted_at IS NULL").
		Where("u.organization_id = ?", organizationID).
		Where("user_profile.job_title > 0").
		Group("jt.name", "jt.priority").
		Order("jt.priority ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

// CountUser : Count total user
func (repo *PgUserRepository) CountUser(organizationID int) (int, error) {
	var total int
	err := repo.DB.Model(&m.User{}).
		ColumnExpr("COUNT(id) AS total").
		Where("organization_id = ?", organizationID).
		Select(&total)

	if err != nil {
		repo.Logger.Error(err)
	}

	return total, err
}

// CountPeopleJobTitle : Count the number of people job title
func (repo *PgUserRepository) CountJapaneseLevel(organizationID int) ([]param.NumberPeopleJpLanguageCert, error) {
	var records []param.NumberPeopleJpLanguageCert
	q := "SELECT elem ->> 'certificate' AS certificate, COUNT(elem ->> 'certificate') AS amount " +
		"FROM user_profiles AS up JOIN users AS u ON u.id = up.user_id, json_array_elements(up.language::json) elem " +
		"WHERE ((u.organization_id = " + strconv.Itoa(organizationID) + ") AND (elem ->> 'language_id' = '2')) " +
		"AND up.deleted_at IS NULL " +
		"GROUP BY elem ->> 'certificate'"

	_, err := repo.DB.Query(&records, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) UpdateIntFieldToNull(organizationID int, field string, valueWhere int) error {
	_, err := repo.DB.Model(&m.UserProfile{}).
		TableExpr("users AS u").
		Set(field+" = NULL").
		Set("updated_at = ?", utils.TimeNowUTC()).
		Where("u.id = user_profile.user_id").
		Where("u.organization_id = ?", organizationID).
		Where("user_profile."+field+" = ?", valueWhere).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgUserRepository) SelectUserIdByEmployeeId(organizationId int, employeeId string) (int, error) {
	var userProfile m.UserProfile
	err := repo.DB.Model(&userProfile).
		Column("user_profile.user_id").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Where("u.organization_id = ?", organizationId).
		Where("user_profile.employee_id = ?", employeeId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return userProfile.UserID, err
}

func (repo *PgUserRepository) SelectEmployeeIdByOrganizationId(organizationId int) ([]param.EmployeeIdAndFullName, error) {
	var records []param.EmployeeIdAndFullName
	err := repo.DB.Model(&m.UserProfile{}).
		Column("user_profile.employee_id", "user_profile.user_id").
		ColumnExpr("user_profile.first_name || ' ' || user_profile.last_name full_name").
		Join("JOIN users AS u ON u.id = user_profile.user_id").
		Where("u.organization_id = ?", organizationId).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) UpdateProfileWithTx(updateProfileParams *[]param.UpdateProfileParams) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		for _, updateParam := range *updateProfileParams {
			_, errTx := tx.Model(&m.UserProfile{}).
				Set("birthday = ?", updateParam.Birthday).
				Set("rank = ?", updateParam.Rank).
				Set("job_title = ?", updateParam.JobTitle).
				Set("phone_number = ?", updateParam.PhoneNumber).
				Set("company_joined_date = ?", updateParam.CompanyJoinedDate).
				Set("branch = ?", updateParam.Branch).
				Set("updated_at = ?", utils.TimeNowUTC()).
				Where("employee_id = ?", updateParam.EmployeeId).
				Update()

			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}
		}

		return errTx
	})

	return err
}

func (repo *PgUserRepository) UpdateLanguageSetting(
	organizationId int,
	languageSettingParams *param.LanguageSettingParams,
) error {
	_, err := repo.DB.Model(&m.User{LanguageId: languageSettingParams.LanguageId}).
		Column("language_id", "updated_at").
		Where("organization_id = ?", organizationId).
		Where("id = ?", languageSettingParams.UserId).
		Update()

	return err
}

func (repo *PgUserRepository) SelectEmailOfGMAndPM(organizationId int) ([]param.EmailOfGMAndPMRecords, error) {
	var records []param.EmailOfGMAndPMRecords

	q := "SELECT DISTINCT u.id AS user_id, u.email, up.first_name || ' ' || up.last_name full_name " +
		"FROM users AS u JOIN user_profiles AS up ON up.user_id = u.id, projects AS p " +
		"WHERE u.organization_id = " + strconv.Itoa(organizationId) + " " +
		"AND (u.role_id = " + strconv.Itoa(cf.GeneralManagerRoleID) + " OR u.id = p.managed_by)" + " " +
		"AND u.deleted_at IS NULL"

	_, err := repo.DB.Query(&records, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) SelectUsersByJpLevel(
	organizationID int,
	jpLevelStatisticDetailParams *param.JpLevelStatisticDetailParams,
) ([]param.FullStatisticDetail, int, error) {
	var totalRow int
	var records []param.FullStatisticDetail

	q1 := "SELECT COUNT(*) " +
		"FROM user_profiles AS user_profile " +
		"JOIN users AS u ON u.id = user_profile.user_id " +
		"LEFT JOIN job_titles AS jt ON jt.id = user_profile.job_title " +
		"LEFT JOIN branches AS b ON b.id = user_profile.branch, json_array_elements(user_profile.language::json) elem " +
		"WHERE ((u.organization_id = " + strconv.Itoa(organizationID) + ") " +
		"AND (elem ->> 'certificate' = '" + jpLevelStatisticDetailParams.Certificate + "')) " +
		"AND user_profile.deleted_at IS NULL"

	_, err := repo.DB.Query(&totalRow, q1)
	if err != nil {
		repo.Logger.Error(err)
	}

	q2 := "SELECT user_profile.user_id, user_profile.birthday, user_profile.company_joined_date, " +
		"user_profile.first_name || ' ' || user_profile.last_name full_name, b.name AS branch, jt.name AS job_title " +
		"FROM user_profiles AS user_profile " +
		"JOIN users AS u ON u.id = user_profile.user_id " +
		"LEFT JOIN job_titles AS jt ON jt.id = user_profile.job_title " +
		"LEFT JOIN branches AS b ON b.id = user_profile.branch, json_array_elements(user_profile.language::json) elem " +
		"WHERE ((u.organization_id = " + strconv.Itoa(organizationID) + ") " +
		"AND (elem ->> 'certificate' = '" + jpLevelStatisticDetailParams.Certificate + "')) " +
		"AND user_profile.deleted_at IS NULL " +
		"ORDER BY user_profile.birthday ASC " +
		"LIMIT " + strconv.Itoa(jpLevelStatisticDetailParams.RowPerPage) + " " +
		"OFFSET " + strconv.Itoa((jpLevelStatisticDetailParams.CurrentPage-1)*jpLevelStatisticDetailParams.RowPerPage)

	_, err = repo.DB.Query(&records, q2)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgUserRepository) SelectIdsOfGMAndManager(organizationId int) ([]int, error) {
	var records []int
	err := repo.DB.Model(&m.User{}).
		Column("id").
		Where("organization_id = ?", organizationId).
		Where("role_id = ?", cf.GeneralManagerRoleID).
		WhereOr("role_id = ?", cf.ManagerRoleID).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) SelectFullNameUser(userId int) (string, error) {
	var fullName string
	err := repo.DB.Model(&m.UserProfile{}).
		ColumnExpr("first_name || ' ' || last_name full_name").
		Where("user_id = ?", userId).
		Select(&fullName)

	if err != nil {
		repo.Logger.Error(err)
	}

	return fullName, err
}

func (repo *PgUserRepository) SelectAvatarUsers(usersId []int) ([]param.UserIdAndAvatarRecord, error) {
	var records []param.UserIdAndAvatarRecord
	err := repo.DB.Model(&m.UserProfile{}).
		Column("user_id", "avatar").
		Where("user_id IN (?)", pg.In(usersId)).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) SelectEmailByUserIds(userIds []int) ([]string, error) {
	var emails []string
	err := repo.DB.Model(&m.User{}).
		Column("email").
		Where("id IN (?)", pg.In(userIds)).
		Select(&emails)

	if err != nil {
		repo.Logger.Error(err)
	}

	return emails, err
}

func (repo *PgUserRepository) SelectIdsOfGM(organizationId int) ([]int, error) {
	var records []int
	err := repo.DB.Model(&m.User{}).
		Column("id").
		Where("organization_id = ?", organizationId).
		Where("role_id = ?", cf.GeneralManagerRoleID).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgUserRepository) GetUserProfileListByOrgId(organizationId int) ([]m.UserExt, error) {
	var userList []m.UserExt

	err := repo.DB.Model(&userList).
		Relation("Role").
		Relation("UserProfile").
		Relation("UserProfile.UserBranch").
		Where("usr.organization_id = ?", organizationId).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return userList, err
}

func (repo *PgUserRepository) UpdateContractExpireDate(userID int, contractExpirationDate time.Time) error {
	userProfile := m.UserProfile{
		ContractExpirationDate: contractExpirationDate,
	}

	_, err := repo.DB.Model(&userProfile).
		Column("contract_expiration_date").
		Where("user_id = ?", userID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}