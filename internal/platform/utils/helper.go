package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
)

// GetSHA256Hash : Convert string(password) to SHA256 string
// Params        : string
// Returns       : sha256
func GetSHA256Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetKeyToken : get secret key from environment variable(declare in docker-compose.yml)
// Params      :
// Returns     : secret key
func GetKeyToken() string {
	keyToken := os.Getenv("KEY_TOKEN")

	return keyToken
}

// PrintVars : use when debug, print object or array to cmd screen
// Params    : any object, array
// Returns   :
// example   : utils.PrintVars(os.Stdout, true, yourObject)
func PrintVars(w io.Writer, writePre bool, vars ...interface{}) {
	if writePre {
		io.WriteString(w, "<pre>\n")
	}
	for i, v := range vars {
		fmt.Fprintf(w, "» item %d type %T:\n", i, v)
		j, err := json.MarshalIndent(v, "", "    ")
		switch {
		case err != nil:
			fmt.Fprintf(w, "error: %v", err)
		case len(j) < 3: // {}, empty struct maybe or empty string, usually mean unexported struct fields
			w.Write([]byte(html.EscapeString(fmt.Sprintf("%+v", v))))
		default:
			w.Write(j)
		}
		w.Write([]byte("\n\n"))
	}
	if writePre {
		io.WriteString(w, "</pre>\n")
	}
}

// GetNameTypeRegistRequests : convert code type of regisration_request to name
// Params      : code type regisration_request
// Returns     : name type regisration_request
func GetNameTypeRegistRequests(typeRegistRequest int) string {
	typeName := ""

	switch typeRegistRequest {
	case cf.MemberSendRequest:
		typeName = "Request"
	case cf.AdminInviteRequest:
		typeName = "Invite"
	}

	return typeName
}

// GetNameStatusRegistRequests : convert code status of regisration_request to name
// Params                      : code status regisration_request
// Returns                     : name status regisration_request
func GetNameStatusRegistRequests(statusRegistRequest int) string {
	statusName := ""

	switch statusRegistRequest {
	case cf.PendingRequestStatus:
		statusName = "Pending"
	case cf.DenyRequestStatus:
		statusName = "Deny"
	case cf.AcceptRequestStatus:
		statusName = "Accepted"
	case cf.RegisteredRequestStatus:
		statusName = "Registered"
	}

	return statusName
}

// GetFieldBindForm : get name error Field when Bind
// Params         : error bind. EX: got=string, field=type, offset=49
// Returns        : value of Field
func GetFieldBindForm(strError string) string {
	var rgx = regexp.MustCompile(`field=(.*?),`)
	arrRgx := rgx.FindStringSubmatch(strError)

	if len(arrRgx) == 0 {
		return ""
	}

	return arrRgx[1]
}

// RemoveIndex : remove element from slice with order
// Params    : inputSlice - slice , index - position
// Returns   : new slice after element removed and keep order
func RemoveIndex(inputSlice []string, index int) []string {
	return append(inputSlice[:index], inputSlice[index+1:]...)
}

// CheckDateBeforeOrEqual : check invalid date from after date to
// Params    : date from, date to, layout format date
// Returns   : new slice after element removed and keep order
func CheckDateBeforeOrEqual(dateFrom string, dateTo string, layoutFormatDate string) bool {
	dateFromTime, _ := time.Parse(layoutFormatDate, dateFrom)
	dateToTime, _ := time.Parse(layoutFormatDate, dateTo)

	if dateFromTime.Before(dateToTime) || dateFromTime.Equal(dateToTime) {
		return true
	}

	return false
}

// GetUniqueString : get unique universal string
// Params    :
// Returns   : unique universal string
func GetUniqueString() string {
	uuid := uuid.New()
	formatUUID := strings.ReplaceAll(uuid.String(), "-", "")
	return formatUUID
}

// TimeNowUTC : get current time with UTC
// Params    : date from, date to, layout format date
// Returns   : new slice after element removed and keep order
func TimeNowUTC() time.Time {
	utc, _ := time.LoadLocation("UTC")
	return time.Now().In(utc)
}

// FindStringInArray : find item in array
// Params    : array, string item
// Returns   : index, bool
func FindStringInArray(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func FindIntInSlice(slice []int, val int) bool {
	if len(slice) == 0 {
		return false
	}

	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}

// GetCurrentQuarterAndYear : Get current quarter and year
// Returns   : int, int
func GetCurrentQuarterAndYear() (int, int) {
	var quarter int
	year := time.Now().Year()

	switch month := int(time.Now().Month()); month {
	case 1, 2, 3:
		quarter = 1
	case 4, 5, 6:
		quarter = 2
	case 7, 8, 9:
		quarter = 3
	default:
		quarter = 4
	}

	return quarter, year
}

// ConvertTwoChar : Convert int to 2 character string
func ConvertTwoChar(number int) string {
	if number < 10 {
		return "0" + strconv.Itoa(number)
	}

	return strconv.Itoa(number)
}

// CompareEqualDate : Compare 2 time is same date
func CompareEqualDate(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(24 * time.Hour).Equal(t2.Truncate(24 * time.Hour))
}

func AppendUniqueSlice(arr1 []int, arr2 []int) []int {
	x := append(arr1, arr2...)
	keys := make(map[int]bool)
	var list []int
	for _, entry := range x {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func IsEqualSlice(a, b []int) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func DiffSlice(prev, next []int) []int {
	var diff []int
	if len(prev) == 0 {
		return next
	}

	for i := 0; i < len(next); i++ {
		if !FindIntInSlice(prev, next[i]) {
			diff = append(diff, next[i])
		}
	}

	return diff
}
