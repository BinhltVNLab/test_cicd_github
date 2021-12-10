package calendar

import (
	"encoding/json"
	"fmt"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Retrieve a token, saves the token, the returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "configs/gcloud/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		panic(err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		panic(err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func InitCalendarService() *calendar.Service {
	b, err := ioutil.ReadFile("configs/gcloud/credentials.json")
	if err != nil {
		panic(err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		panic(err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		panic(err)
	}

	return srv
}

func AddLeaveEvent(leaveType int, summary string, description string, start string, end string) *calendar.Event {
	srv := InitCalendarService()

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Reminders: &calendar.EventReminders{
			ForceSendFields: []string{"UseDefault"},
		},
	}

	switch leaveType {
	case cf.FullDayOff, cf.MorningOff, cf.AfternoonOff:
		event.Start = &calendar.EventDateTime{
			Date:     start,
			TimeZone: "Asia/Ho_Chi_Minh",
		}
		event.End = &calendar.EventDateTime{
			Date:     end,
			TimeZone: "Asia/Ho_Chi_Minh",
		}
	case cf.LateForWork, cf.LeaveEarly, cf.GoOutside:
		event.Start = &calendar.EventDateTime{
			DateTime: start,
			TimeZone: "Asia/Ho_Chi_Minh",
		}
		event.End = &calendar.EventDateTime{
			DateTime: end,
			TimeZone: "Asia/Ho_Chi_Minh",
		}
	default:
		if strings.Contains(start, "T") {
			event.Start = &calendar.EventDateTime{
				DateTime: start,
				TimeZone: "Asia/Ho_Chi_Minh",
			}
			event.End = &calendar.EventDateTime{
				DateTime: end,
				TimeZone: "Asia/Ho_Chi_Minh",
			}
		} else {
			event.Start = &calendar.EventDateTime{
				Date:     start,
				TimeZone: "Asia/Ho_Chi_Minh",
			}
			event.End = &calendar.EventDateTime{
				Date:     end,
				TimeZone: "Asia/Ho_Chi_Minh",
			}
		}
	}

	event, err := srv.Events.Insert(cf.CalendarId, event).Do()
	if err != nil {
		panic(err)
	}

	return event
}

func AddEventRemind(summary string, description string, start string, end string) *calendar.Event {
	srv := InitCalendarService()

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Reminders: &calendar.EventReminders{
			ForceSendFields: []string{"UseDefault"},
		},
	}

	event.Start = &calendar.EventDateTime{
		Date:     start,
		TimeZone: "Asia/Ho_Chi_Minh",
	}
	event.End = &calendar.EventDateTime{
		Date:     end,
		TimeZone: "Asia/Ho_Chi_Minh",
	}

	event, err := srv.Events.Insert(cf.CalendarId, event).Do()
	if err != nil {
		panic(err)
	}

	return event
}

func RemoveLeaveEvent(eventId string) {
	srv := InitCalendarService()
	err := srv.Events.Delete(cf.CalendarId, eventId).Do()
	if err != nil {
		panic(err)
	}
}
