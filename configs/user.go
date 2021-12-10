package configs


var EnUserCategories = map[string]string{
	"Employee Id":        "Employee Id",
	"Full Name":          "Full Name",
	"Company join date": "Entering company",
	"Branch":             "Branch",
	"Birthday": "Birthday",
	"Phone number":"Phone number",
	"Job title": "Job title",
	"Gender": "Gender",
	"Email": "Email",
}

var JpUserCategories = map[string]string{
	"Employee Id":        "社員コード",
  	"Company join date": "入社日" ,
	"Birthday": "生年月日",
	"Phone number":"電話番号",
	"Job title": "職名",
	"Gender": "性別",
	"Email": "メール",
	"Full Name":          "氏名",
	"Branch":             "支店",
}

var VnUserCategories = map[string]string{
	"Employee Id":        "Mã nhân viên",
	"Full Name":          "Họ và tên",
	"Company join date":		"Ngày vào công ty",
	"Branch":             "Chi nhánh",
	"Birthday": "Sinh Nhật",
	"Phone number":"Số điện thoại",
	"Job title": "Chức vụ",
	"Gender": "Giới tính",
	"Email": "Email",
}
var GenderExcel = map[int]string{
	1: "Nam",
	2: "Nữ",
}
var JobTitleExcel = map[int]string{
	1 : "Director",
  2 : "Manager",
  3 : "Assistant manager",
  4 : "Engineer",
  5 : "Assistant",
  6 : "Part time",
  7 : "Intern",
  8 : "Vice Director",
  9 : "Leader",
  10 : "Designer",
  11 : "Other",
}
var BranchExcel = map[int]string{
	1: "Hồ Chí Minh",
	2: "Hà Nội",
	3: "Đà Nẵng",
	4: "Tokyo",
	5: "Osaka",
}
