package configs

// Common const
const (
	SuccessRequestTemplate      = "internal/platform/email/template/successRequest.html"
	CreateOrganizationTemplate  = "internal/platform/email/template/createOrganization.html"
	ChangeEmailTemplate         = "internal/platform/email/template/changeEmail.html"
	RegisterSuccessfulTemplate  = "internal/platform/email/template/registerSuccessfulTemplate.html"
	TemplateSendMailAnnounce    = "internal/platform/email/template/templateAnnounce.html"
	TemplateMailForgotPass      = "internal/platform/email/template/forgotPassword.html"
	LeaveRequestTemplate        = "internal/platform/email/template/leaveRequest.html"
	OvertimeRequestTemplate     = "internal/platform/email/template/overtimeRequest.html"
	SendTestMailTemplate        = "internal/platform/email/template/sendTestMail.html"
	Recruitment                 = "internal/platform/email/template/recruitment.html"
	DirectoryAvatarImage        = "internal/platform/cloud/images/"
	ExpiredHours                = 2                     // expired time for registration code , for now 2 hours
	FormatDate                  = "2006-01-02 15:04:05" // must be set this format for parse date
	FormatDateNoSec             = "2006-01-02 15:04"
	FormatDateDisplay           = "2006/01/02"
	FormatDateDatabase          = "2006-01-02"
	PostgreCharacterDisplayDate = "YYYY/MM/DD" // just use for Postgresql
	FormatTimeDisplay           = "2006/01/02 15:04"
	RowPerPageProfileList       = 10
	AvatarFolderGCS             = "images/avatar/"
	ContractTypeFolderGCS       = "contract_types/"
	CVFOLDERGCS                 = "cvs/"
	CONTRACTFOLDERGCS           = "contracts/"
	FormatDisplayTimekeeping    = "2006/01/02 15:04 PM"
	FormatDisplayTimekeeping2   = "2006/01/02 15:04:05"
	EnLanguageId                = 1
	JpLanguageId                = 2
	VnLanguageId                = 3

	ORGANIZATIONEMAILSETTING = 1
	BRANCHSETTING            = 2
	JOBTITLESETTING          = 3
	TECHNOLOGYSETTING        = 4
	OVERTIMESETTING          = 5
	FINISHSETTING            = 6

	Trainee = 1
	Junior  = 2
	Middle  = 3
	Senior  = 4

	RecruitmentModule = 0
	SuerveyMoldule    = 1
	RequestModule     = 2
	TodolistModule    = 3

	TOTALMODULE = 5
	ACCESSFUNC = 1
	NOTACCESSFUNC = 2

	JOBACTIVE = 1
	JOBDONE = 2

	AcceptRequestBorrowAsset = 1
	AcceptRequestReturnAsset = 2
	PendingRequestBorrowAsset = 3
	PendingRequestReturnAsset = 4
	DenyRequestBorrow = 5
	DenyRequestReturn = 6
)

// JpLanguageLevel map
var JpLanguageLevel = map[int]string{
	1: "N1",
	2: "N2",
	3: "N3",
	4: "N4",
	5: "N5",
	6: "Japanese people",
}

// EvaluationRankList list of Rank
// TODO : create story to save this in database
var EvaluationRankList = map[int]string{
	1: "D",
	2: "A",
	3: "B",
	4: "C",
	5: "S",
}

// QuarterList list of branch
// TODO : create story to save this in database
var QuarterList = []int{1, 2, 3, 4}

// RankList list of Rank
// TODO : create story to save this in database
var RankList = map[int]string{
	10: "1-0",
	11: "1-1",
	12: "1-2",
	13: "1-3",
	20: "2-0",
	21: "2-1",
	22: "2-2",
	23: "2-3",
	24: "2-4",
	25: "2-5",
	26: "2-6",
	30: "3-0",
	31: "3-1",
	32: "3-2",
	33: "3-3",
	34: "3-4",
	40: "4-0",
	41: "4-1",
	42: "4-2",
	43: "4-3",
	50: "5-0",
	51: "5-1",
	52: "5-2",
	53: "5-3",
	60: "6-0",
	61: "6-1",
	62: "6-2",
	63: "6-3",
}
var WorkingStatus = map[int]string{
	1: "Working",
	2: "Onsite",
	3: "Severance",
}

//LanguageList list language
var LanguageList = map[int]string{
	1: "English",
	2: "Japanese",
	3: "Vietnamese",
}

// LevelLanguageList level language list
var LevelLanguageList = map[int]string{
	1: "Beginning",
	2: "Immediate",
	3: "Advanced",
}

// AllowFormatImageList format image allow
var AllowFormatImageList = []string{
	"png",
	"jpg",
	"jpeg",
	"gif",
}

var LevelSkillList = map[int]string{
	Trainee: "Trainee",
	Junior:  "Junior",
	Middle:  "Middle",
	Senior:  "Senior",
}

var PermissionStatus = map[int]bool{
	1: true,
	2: false,
}

// Status of asset
var AssetStatus = map[int]string{
	1: "Free",
	2: "Using",
	3: "Sold",
	4: "Guaranteeing",
	5: "Broken",
	6: "Lost",
	7: "Other",
}

// RequestAssetStatus Status of request asset
var RequestAssetStatus = map[int]string{
	1: "Accept",
	2: "Reject",
	3: "Pending",
}

// ModulesOrg : List of modules
var ModulesOrg = `[
	{
	"module_id": 1,
	"status": true
	},
	{
	"module_id": 2,
	"status": true
	},
	{
	"module_id": 3,
	"status": true
	},
	{
	"module_id": 4,
	"status": true
	},
	{
	"module_id": 5,
	"status": true
	}
]`