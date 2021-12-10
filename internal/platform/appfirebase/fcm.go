package appfirebase

import (
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"os"
)

type FirebaseCloudMessage struct {
	ctx    context.Context
	client *messaging.Client
}

func (f *FirebaseCloudMessage) InitFcm() {
	opt := option.WithCredentialsFile("configs/gcloud/firebaseServiceAccount.json")
	config := &firebase.Config{ProjectID: "microerp-265008"}
	f.ctx = context.Background()
	app, err := firebase.NewApp(f.ctx, config, opt)
	if err != nil {
		panic(err)
	}

	client, err := app.Messaging(f.ctx)
	if err != nil {
		panic(err)
	}

	f.client = client
}

func (f *FirebaseCloudMessage) SendMessageToSpecificUser(registrationToken string, title string, body string, link string) error {
	notification := new(messaging.Notification)
	notification.Title = title
	notification.Body = body

	fcmOption := new(messaging.WebpushFcmOptions)
	fcmOption.Link = os.Getenv("BASE_SPA_URL") + link

	webpush := new(messaging.WebpushConfig)
	webpush.FcmOptions = fcmOption

	data := map[string]string{
		"link": link,
	}

	message := &messaging.Message{
		Notification: notification,
		Webpush:      webpush,
		Token:        registrationToken,
		Data:         data,
	}

	_, err := f.client.Send(f.ctx, message)
	return err
}

func (f *FirebaseCloudMessage) SendMessageToMultiUsers(registrationTokens []string, title string, body string, link string) {
	notification := new(messaging.Notification)
	notification.Title = title
	notification.Body = body

	fcmOption := new(messaging.WebpushFcmOptions)
	fcmOption.Link = os.Getenv("BASE_SPA_URL") + link

	webpush := new(messaging.WebpushConfig)
	webpush.FcmOptions = fcmOption

	data := map[string]string{
		"link": link,
	}

	message := &messaging.MulticastMessage{
		Notification: notification,
		Webpush:      webpush,
		Tokens:       registrationTokens,
		Data:         data,
	}

	_, err := f.client.SendMulticast(f.ctx, message)
	if err != nil {
		panic(err)
	}
}
