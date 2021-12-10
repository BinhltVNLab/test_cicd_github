package configs

const (
	DayOffTypeOvertime = 1
	MoneyTypeOvertime  = 2

	WorkAtNoon    = 1
	NotWorkAtNoon = 2
)

var MapOvertimeType = map[int]string{
	DayOffTypeOvertime: "Take Day Off",
	MoneyTypeOvertime:  "Take Money",
}

var VnMapOvertimeType = map[int]string{
	DayOffTypeOvertime: "Lấy ngày nghỉ",
	MoneyTypeOvertime:  "Lấy tiền",
}

var JpMapOvertimeType = map[int]string{
	DayOffTypeOvertime: "有給休暇交換",
	MoneyTypeOvertime:  "お金",
}

var MapStatusOvertimeType = map[int]string{
	PendingRequestStatus: "Pending",
	DenyRequestStatus:    "Deny",
	AcceptRequestStatus:  "Accept",
}

var VnMapStatusOvertimeType = map[int]string{
	PendingRequestStatus: "Đang duyệt",
	DenyRequestStatus:    "Từ chối",
	AcceptRequestStatus:  "Chấp nhận",
}

var JpMapStatusOvertimeType = map[int]string{
	PendingRequestStatus: "保留中",
	DenyRequestStatus:    "否定",
	AcceptRequestStatus:  "同意",
}

var VnWeekDay = map[string]string{
	"1" : "Thứ 2",
	"2" : "Thứ 3",
	"3" : "Thứ 4",
	"4" : "Thứ 5",
	"5" : "Thứ 6",
	"6" : "Thứ 7",
	"7" : "Chủ nhật",
}

var EnWeekDay = map[string]string{
	"1" : "Monday",
	"2" : "Tuesday",
	"3" : "Wednesday",
	"4" : "Thursday",
	"5" : "Friday",
	"6" : "Saturday",
	"7" : "Sunday",
}

var JpWeekDay = map[string]string{
	"1" : "月曜日",
	"2" : "火曜日",
	"3" : "水曜日",
	"4" : "木曜日",
	"5" : "金曜日",
	"6" : "土曜日",
	"7" : "日曜日",
}

var EnOvertimeCategories = map[string]string{
	"Employee Id":        "Employee Id",
	"Full Name":          "Full Name",
	"Branch":             "Branch",
	"Project":            "Project",
	"Date":               "Date",
	"Weekday":            "Weekday",
	"Range Time":         "Range Time",
	"Working Time":       "Working Time",
	"Weight":             "Weight",
	"Total Working Time": "Total Working Time",
	"Type":               "Type",
	"Status":             "Status",
	"Note":               "Note",
}

var JpOvertimeCategories = map[string]string{
	"Employee Id":        "社員コード",
	"Full Name":          "氏名",
	"Branch":             "支店",
	"Project":            "プロジェクト",
	"Date":               "残業日",
	"Weekday":            "曜日",
	"Range Time":         "残業期間",
	"Working Time":       "残業時間",
	"Weight":             "ウェイト",
	"Total Working Time": "合計残業時間",
	"Type":               "残業タイプ",
	"Status":             "承認状況",
	"Note":               "備考",
}

var VnOvertimeCategories = map[string]string{
	"Employee Id":        "Mã nhân viên",
	"Full Name":          "Họ và tên",
	"Branch":             "Chi nhánh",
	"Project":            "Dự án",
	"Date":               "Ngày",
	"Weekday":            "Thứ",
	"Range Time":         "Khoảng thời gian",
	"Working Time":       "Thời gian làm việc",
	"Weight":             "Trọng số",
	"Total Working Time": "Thời gian làm thực",
	"Type":               "Loại",
	"Status":             "Trạng thái",
	"Note":               "Ghi chú",
}
