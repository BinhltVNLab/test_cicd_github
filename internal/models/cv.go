package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

type Cv struct {
	cm.BaseModel
	tableName       struct{} `sql:"alias:c"`
	RecruitmentId   int
	MediaId         int
	FileName        string
	FullName        string
	PhoneNumber     string
	Email           string
	Salary          string
	DateReceiptCv   time.Time
	InterviewMethod int
	ContactLink     string
	StatusCv        int
	MediaIdOther	string
}

type LogCvState struct {
	cm.BaseModel
	tableName struct{} `sql:"alias:lcs"`

	CvId      int
	Status    int
	UpdateDay time.Time
}
