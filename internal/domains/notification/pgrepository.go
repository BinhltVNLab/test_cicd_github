package notification

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
	"strconv"
	"time"
)

type PgNotificationRepository struct {
	cm.AppRepository
}

func NewPgNotificationRepository(logger echo.Logger) (repo *PgNotificationRepository) {
	repo = &PgNotificationRepository{}
	repo.Init(logger)
	return
}

func (repo *PgNotificationRepository) InsertNotificationWithTx(tx *pg.Tx, organizationId int, sender int, params *param.InsertNotificationParam) error {
	notification := m.Notification{
		OrganizationId: organizationId,
		Sender:         sender,
		Receiver:       params.Receiver,
		Content:        params.Content,
		Status:         cf.NotificationStatusUnread,
		RedirectUrl:    params.RedirectUrl,
	}

	err := tx.Insert(&notification)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgNotificationRepository) UpdateNotificationStatusRead(organizationId int, receiver int) error {
	q := "UPDATE notifications AS n " +
		"SET status = " + strconv.Itoa(cf.NotificationStatusRead) + ", updated_at = '" + utils.TimeNowUTC().Format(cf.FormatDate) + "' " +
		"FROM (" +
		"SELECT ntf.id " +
		"FROM notifications AS ntf " +
		"WHERE ntf.organization_id = " + strconv.Itoa(organizationId) + " " +
		"AND ntf.receiver = " + strconv.Itoa(receiver) + " " +
		"AND ntf.status = " + strconv.Itoa(cf.NotificationStatusUnread) + " " +
		"AND EXTRACT(DAY FROM (CURRENT_TIMESTAMP - ntf.created_at)) < 31 " +
		"AND ntf.deleted_at IS NULL) nt " +
		"WHERE n.id = nt.id"

	_, err := repo.DB.Query(&m.Notification{}, q)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgNotificationRepository) UpdateNotificationStatus(params *param.UpdateNotificationStatusParam) error {
	notification := m.Notification{Status: params.Status}
	if params.Status == cf.NotificationStatusSeen {
		notification.DatetimeSeen = utils.TimeNowUTC()
	} else {
		notification.DatetimeSeen = time.Time{}
	}

	_, err := repo.DB.Model(&notification).
		Column("status", "datetime_seen", "updated_at").
		Where("id = ?", params.Id).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgNotificationRepository) CheckNotificationExist(id int) (int, error) {
	count, err := repo.DB.Model(&m.Notification{}).Where("id = ?", id).Count()
	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgNotificationRepository) SelectNotifications(
	organizationId int,
	params *param.GetNotificationsParam,
) ([]param.GetNotificationRecord, int, error) {
	var records []param.GetNotificationRecord
	totalRow, err := repo.DB.Model(&m.Notification{}).
		Column("ntf.id", "ntf.content", "ntf.status", "ntf.redirect_url", "ntf.created_at").
		ColumnExpr("up.first_name || ' ' || up.last_name sender").
		ColumnExpr("up.avatar AS avatar_sender").
		Join("JOIN user_profiles AS up ON up.user_id = ntf.sender").
		Where("ntf.organization_id = ?", organizationId).
		Where("ntf.receiver = ?", params.Receiver).
		Order("ntf.created_at DESC").
		Offset((params.CurrentPage - 1) * params.RowPerPage).
		Limit(params.RowPerPage).
		SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgNotificationRepository) CountNotificationsUnRead(organizationId int, receiver int) (int, error) {
	count, err := repo.DB.Model(&m.Notification{}).
		Where("organization_id = ?", organizationId).
		Where("receiver = ?", receiver).
		Where("status = ?", cf.NotificationStatusUnread).
		Where("EXTRACT(DAY FROM (CURRENT_TIMESTAMP - created_at)) < 31").
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgNotificationRepository) DeleteNotification(id int) error {
	_, err := repo.DB.Model(&m.Notification{}).Where("id = ?", id).Delete()
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgNotificationRepository) InsertEmailNotiRequest(emailNotiRequest *m.EmailNotiRequest) error {
	return repo.DB.Insert(emailNotiRequest)
}

func (repo *PgNotificationRepository) FindNotiRequestById(id int) (*m.NotiRequest, error) {
	notiRequest := m.NotiRequest{}
	err := repo.DB.Model(&notiRequest).Where("id = ?", id).Select()
	if err != nil {
		repo.Logger.Error(err)
	}

	return &notiRequest, err
}

func (repo *PgNotificationRepository) UpdateNotiRequestStatus(id int, fromStatus string, toStatus string) error {
	result, err := repo.DB.Model(&m.NotiRequest{}).
		Where("id = ? and status = ?", id, fromStatus).
		Set("status = ?", toStatus).
		Update()

	if err != nil {
		repo.Logger.Error(err)
		return err
	}

	if result.RowsAffected() < 1 {
		return fmt.Errorf("update noti-request id = %v and status = %v has no row affected", id, toStatus)
	}

	return nil
}

func (repo *PgNotificationRepository) GetNotificationsByNotiRequest(notiRequest *m.NotiRequest) ([]*param.GetNotificationByNotiRequestRecord, error) {
	var notifications []*param.GetNotificationByNotiRequestRecord
	err := repo.DB.Model(&m.Notification{}).
		Column("ntf.id", "ntf.sender", "ntf.title", "ntf.content", "ntf.redirect_url", "ntf.receiver").
		ColumnExpr("up.first_name || ' ' || up.last_name sender_name").
		Join("JOIN user_profiles AS up ON up.user_id = ntf.sender").
		Where("ntf.noti_request_id = ?", notiRequest.ID).
		Order("ntf.created_at DESC").
		Select(&notifications)

	if err != nil {
		repo.Logger.Error(err)
	}

	return notifications, err
}

func (repo *PgNotificationRepository) GetEmailNotificationsByNotiRequest(notiRequest *m.NotiRequest) ([]*param.GetEmailNotiRequestByNotiRequestRecord, error) {
	var emailNotiRequests []*param.GetEmailNotiRequestByNotiRequestRecord
	err := repo.DB.Model(&m.EmailNotiRequest{}).
		Column("enr.id", "enr.sender", "enr.to_user_ids", "enr.subject", "enr.content", "enr.url", "enr.template").
		Column("org.email", "org.email_password").
		Join("JOIN organizations AS org ON enr.organization_id = org.id").
		Where("enr.noti_request_id = ?", notiRequest.ID).
		Order("enr.created_at DESC").
		Select(&emailNotiRequests)

	if err != nil {
		repo.Logger.Error(err)
	}

	return emailNotiRequests, err
}
