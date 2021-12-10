package repository

import (
	"github.com/go-pg/pg/v9"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type NotificationRepository interface {
	InsertNotificationWithTx(tx *pg.Tx, organizationId int, sender int, params *param.InsertNotificationParam) error
	UpdateNotificationStatusRead(organizationId int, receiver int) error
	UpdateNotificationStatus(params *param.UpdateNotificationStatusParam) error
	CheckNotificationExist(id int) (int, error)
	SelectNotifications(organizationId int, params *param.GetNotificationsParam) ([]param.GetNotificationRecord, int, error)
	CountNotificationsUnRead(organizationId int, receiver int) (int, error)
	DeleteNotification(id int) error
	InsertEmailNotiRequest(*m.EmailNotiRequest) error
	FindNotiRequestById(id int) (*m.NotiRequest, error)
	UpdateNotiRequestStatus(id int, fromStatus string, toStatus string) error
	GetNotificationsByNotiRequest(request *m.NotiRequest) ([]*param.GetNotificationByNotiRequestRecord, error)
	GetEmailNotificationsByNotiRequest(notiRequest *m.NotiRequest) ([]*param.GetEmailNotiRequestByNotiRequestRecord, error)
}
