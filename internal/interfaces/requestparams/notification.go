package requestparams

import "time"

type InsertNotificationParam struct {
	Receiver    int    `json:"receiver" valid:"required"`
	Content     string `json:"content" valid:"required"`
	RedirectUrl string `json:"redirect_url" valid:"required"`
}

type UpdateNotificationStatusReadParam struct {
	Receiver int `json:"receiver" valid:"required"`
}

type GetNotificationsParam struct {
	Receiver    int `json:"receiver" valid:"required"`
	CurrentPage int `json:"current_page" valid:"required"`
	RowPerPage  int `json:"row_per_page"`
}

type GetNotificationRecord struct {
	Id           int       `json:"id"`
	Sender       string    `json:"sender"`
	AvatarSender string    `json:"avatar_sender"`
	Content      string    `json:"content"`
	Status       int       `json:"status"`
	RedirectUrl  string    `json:"redirect_url"`
	CreatedAt    time.Time `json:"created_at"`
}

type GetNotificationByNotiRequestRecord struct {
	Id          int
	Sender      int
	Receiver    int
	Content     string
	RedirectUrl string
	Title       string
	SenderName  string
}

type GetEmailNotiRequestByNotiRequestRecord struct {
	Id            int
	Sender        int
	ToUserIds     []int `pg:",array"`
	Subject       string
	Content       string
	Url           string
	Template      string

	Email         string
	EmailPassword string
}

type RemoveNotificationParam struct {
	Id       int `json:"id" valid:"required"`
	Receiver int `json:"receiver" valid:"required"`
}

type UpdateNotificationStatusParam struct {
	Id       int `json:"id" valid:"required"`
	Status   int `json:"status" valid:"required"`
	Receiver int `json:"receiver" valid:"required"`
}

type GetTotalNotificationsUnreadParam struct {
	ClientTime string `json:"client_time" valid:"required"`
}

type SampleData struct {
	URL     string
	SendTo  []string
	SendCc  []string
	Content string
	OrgTag  string
}

type SendNotiRequestParam struct {
	Id int `json:"id" valid:"required"`
}
