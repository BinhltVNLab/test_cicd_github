package configs

const (
	NotificationStatusUnread = 1
	NotificationStatusRead   = 2
	NotificationStatusSeen   = 3
)

var NotificationStatusMap = map[int]string {
	NotificationStatusUnread: "Unread",
	NotificationStatusRead: "Read",
	NotificationStatusSeen: "Seen",
}
