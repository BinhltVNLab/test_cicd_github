package configs

// Leave
const (
	// Leave type
	FullDayOff   = 1
	MorningOff   = 2
	AfternoonOff = 3
	LateForWork  = 4
	LeaveEarly   = 5
	GoOutside    = 6
	WorkAtHome   = 7
	BusinessTrip = 8
	OtherLeave   = 9

	// Leave bonus type
	AnnualLeave      = 1
	SeniorityLeave   = 2
	SickLeave        = 3
	MarryLeave       = 4
	MaternityLeave   = 5
	BereavementLeave = 6
	ClearLeave       = 7
	OvertimeLeave    = 8

	// Subtract day off type
	Subtract  = 1
	ExtraWork = 2
	Event     = 3
	Other     = 4

	BreakLunchStart = "12:00"
	BreakLunchEnd   = "13:30"

	NoticeEmail = "notice@vietnamlab.vn"
	CalendarId  = "runsystem.net_ks1mp6gdm9c73i3r9a18n66d7k@group.calendar.google.com"
)

// LeaveRequestTypes map
var LeaveRequestTypes = map[int]string{
	FullDayOff:   "Full day off",
	MorningOff:   "Morning off",
	AfternoonOff: "Afternoon off",
	LateForWork:  "Late for work",
	LeaveEarly:   "Leave early",
	GoOutside:    "Go outside",
	WorkAtHome:   "Work at home",
	BusinessTrip: "Business trip",
	OtherLeave:   "Other leave",
}

// LeaveBonusTypes map
var LeaveBonusTypes = map[int]string{
	AnnualLeave:      "Annual leave",
	SeniorityLeave:   "Seniority leave",
	SickLeave:        "Sick leave",
	MarryLeave:       "Marry leave",
	MaternityLeave:   "Maternity leave",
	BereavementLeave: "Bereavement leave",
	ClearLeave:       "Clear leave",
	OvertimeLeave:    "Overtime leave",
}

// LeaveRequestJpTypes map
var LeaveRequestJpTypes = map[int]string{
	FullDayOff:   "全休",
	MorningOff:   "午前休",
	AfternoonOff: "午後休",
	LateForWork:  "遅刻",
	LeaveEarly:   "早退",
	GoOutside:    "外出",
	WorkAtHome:   "自宅業務",
	BusinessTrip: "出張",
	OtherLeave:   "その他の",
}

var SubtractDayOffTypes = map[int]string{
	Subtract:  "Subtract",
	ExtraWork: "Extra Work",
	Event:     "Go Event",
	Other:     "Other",
}

var LeaveCategoriesEn = map[string]string{
	"Employee Id":        "Employee Id",
	"Full Name":          "Full Name",
	"Leave Request Type": "Leave Request Type",
	"Date From":          "Date From",
	"Date To":            "Date To",
	"Time":               "Time",
	"Note":               "Note",
}

var LeaveCategoriesJp = map[string]string{
	"Employee Id":        "社員番号",
	"Full Name":          "名",
	"Leave Request Type": "タイプ",
	"Date From":          "時間から",
	"Date To":            "時間まで",
	"Time":               "タイム",
	"Note":               "ノート",
}

var LeaveCategoriesVn = map[string]string{
	"Employee Id":        "Mã nhân viên",
	"Full Name":          "Họ và tên",
	"Leave Request Type": "Kiểu xin nghỉ",
	"Date From":          "Ngày bắt đầu",
	"Date To":            "Ngày kết thúc",
	"Time":               "Thời gian",
	"Note":               "Ghi chú",
}
