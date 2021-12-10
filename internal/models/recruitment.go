package models

import (
	"strconv"
	"strings"
	"time"

	"gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type Recruitment struct {
	cm.BaseModel

	tableName         struct{} `sql:"alias:rec"`
	OrganizationId    int
	JobName           string
	StartDate         time.Time
	ExpiryDate        time.Time
	BranchIds         []int `pg:",array"`
	Assignees         []int `pg:",array"`
	DetailJobFileName string
}

func (m *Recruitment) IsAllowCRUD_CV(loggedInUser User) bool {
	return loggedInUser.OrganizationID == m.OrganizationId ||
		(loggedInUser.RoleID == configs.GeneralManagerRoleID &&
			utils.FindIntInSlice(m.Assignees, loggedInUser.ID))
}

func (m *Recruitment) GCSCVDirectory() string {
	return configs.CVFOLDERGCS + strconv.Itoa(m.OrganizationId) + "/" +
		strings.Replace(m.JobName, " ", "_", -1)
}

type DetailedJobRecruitment struct {
	cm.BaseModel

	tableName         struct{} `sql:"alias:djr"`
	RecruitmentID     int
	Amount            int
	Address           []string `pg:",array"`
	Place             []string `pg:",array"`
	Role              int
	Gender            int
	TypeOfWork        string
	Experience        int
	SalaryType        int
	SalaryFrom        int
	SalaryTo          int
	ProfileRecipients string
	Email             string
	PhoneNumber       string
	Description       string

	Recruitment *Recruitment `pg:",fk:recruitment_id"`
}

func (m *Recruitment) IsAllowCUD_DetailedJob(loggedInUser User) bool {
	return loggedInUser.OrganizationID == m.OrganizationId ||
		(loggedInUser.RoleID == configs.GeneralManagerRoleID &&
			utils.FindIntInSlice(m.Assignees, loggedInUser.ID))
}
