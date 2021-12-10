package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	"time"
)

type Notification struct {
	cm.BaseModel

	tableName      struct{} `sql:"alias:ntf"`
	OrganizationId int
	Sender         int
	Receiver       int
	Content        string
	DatetimeSeen   time.Time
	RedirectUrl    string
	Status         int
	Title          string
	NotiRequestId  int
}

const (
	NotiRequestStatusInitial          = "INITIAL"
	NotiRequestStatusProcessing       = "PROCESSING"
	NotiRequestStatusSucceedProcessed = "SUCCEED"
	NotiRequestStatusFailedProcessed  = "FAILED"
)

type NotiRequest struct {
	cm.BaseModel

	tableName struct{} `sql:"alias:ntr"`
	Status    string
}

type EmailNotiRequest struct {
	cm.BaseModel

	tableName      struct{} `sql:"alias:enr"`
	NotiRequestId  int
	OrganizationId int
	Sender         int
	ToUserIds      []int `pg:",array"`
	Subject        string
	Content        string
	Url            string
	Template       string
}
