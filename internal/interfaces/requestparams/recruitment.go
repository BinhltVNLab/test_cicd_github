package requestparams

import (
	"time"
)

type CreateJobParam struct {
	JobName    string `json:"job_name" valid:"required"`
	StartDate  string `json:"start_date" valid:"required"`
	ExpiryDate string `json:"expiry_date" valid:"required"`
	BranchIds  []int  `json:"branch_ids" valid:"required"`
	Assignees  []int  `json:"assignees" valid:"required"`
}

type CreateDetailJobParam struct {
	RecruitmentId     int      `json:"recruitment_id" valid:"required"`
	Amount            int      `json:"amount" valid:"required"`
	Place             []string `json:"place" valid:"required"`
	Address           []string `json:"address" valid:"required"`
	Role              int      `json:"role" valid:"required"`
	Gender            int      `json:"gender"`
	TypeOfWork        int      `json:"type_of_work" valid:"required"`
	Experience        int      `json:"experience" valid:"required"`
	SalaryType        int      `json:"salary_type" valid:"required"`
	SalaryFrom        int      `json:"salary_from"`
	SalaryTo          int      `json:"salary_to"`
	ProfileRecipients string   `json:"profile_recipients"`
	Email             string   `json:"email"`
	PhoneNumber       string   `json:"phone_number"`
	Description       string   `json:"description"`
}

type UpdateDetailJobParam struct {
	Id                int      `json:"id" valid:"required"`
	Amount            int      `json:"amount" valid:"required"`
	Place             []string `json:"place" valid:"required"`
	Address           []string `json:"address" valid:"required"`
	Role              int      `json:"role" valid:"required"`
	Gender            int      `json:"gender"`
	TypeOfWork        int      `json:"type_of_work" valid:"required"`
	Experience        int      `json:"experience" valid:"required"`
	SalaryType        int      `json:"salary_type" valid:"required"`
	SalaryFrom        int      `json:"salary_from"`
	SalaryTo          int      `json:"salary_to"`
	ProfileRecipients string   `json:"profile_recipients"`
	Email             string   `json:"email"`
	PhoneNumber       string   `json:"phone_number"`
	Description       string   `json:"description"`
}

type UploadCvParam struct {
	RecruitmentId int       `json:"recruitment_id"`
	CvFields      []CvField `json:"cv_fields"`
	Assignees     []int     `json:"assignees"`
}

type CreateCvParam struct {
	RecruitmentId   int    `json:"recruitment_id" valid:"required"`
	FullName        string `json:"full_name" valid:"required"`
	PhoneNumber     string `json:"phone_number" valid:"required"`
	Email           string `json:"email" valid:"required,email"`
	Salary          string `json:"salary"`
	DateReceiptCv   string `json:"date_receipt_cv" valid:"required"`
	InterviewMethod int    `json:"interview_method"`
	ContactLink     string `json:"contact_link"`
	MediaIdOther     string `json:"media_id_other"`
	MediaId         int    `json:"media_id"`
	FileName        string `json:"file_name" valid:"required"`
	FileContent     string `json:"file_content" valid:"required"`
	Assignees     []int     `json:"assignees"`
}

type UpdateCvParam struct {
	Id              int    `json:"id" valid:"required"`
	RecruitmentId   int    `json:"recruitment_id" valid:"required"`
	FullName        string `json:"full_name" valid:"required"`
	PhoneNumber     string `json:"phone_number" valid:"required"`
	Email           string `json:"email" valid:"required,email"`
	Salary          string `json:"salary"`
	DateReceiptCv   string `json:"date_receipt_cv" valid:"required"`
	InterviewMethod int    `json:"interview_method"`
	ContactLink     string `json:"contact_link"`
	MediaIdOther     string `json:"media_id_other"`
	MediaId         int    `json:"media_id"`
	FileName        string `json:"file_name"`
	FileContent     string `json:"file_content"`
}

type GetCVByIdParam struct {
	Id int `json:"id" valid:"required"`
}

type CvField struct {
	MediaIdOther     string `json:"media_id_other"`
	MediaId  int    `json:"media_id"`
	FileName string `json:"file_name"`
	Content  string `json:"content"`
	Status   int    `json:"status"`
}

type EditJobParam struct {
	Id         int    `json:"id" valid:"required"`
	JobName    string `json:"job_name" valid:"required"`
	StartDate  string `json:"start_date" valid:"required"`
	ExpiryDate string `json:"expiry_date"`
	BranchIds  []int  `json:"branch_ids" valid:"required"`
	Assignees  []int  `json:"assignees" valid:"required"`
}

type GetJobsParam struct {
	JobName     string `json:"job_name"`
	JobStatus   int    `json:"job_status"`
	ExpiryDate  string `json:"expiry_date"`
	BranchId    int    `json:"branch_id"`
	UserID      int    `json:"user_id"`
	CurrentPage int    `json:"current_page" valid:"required"`
	RowPerPage  int    `json:"row_per_page" valid:"required"`
}

type GetJobParam struct {
	Id int `json:"id" valid:"required"`
}

type GetJobRecords struct {
	Id         int       `json:"id"`
	JobName    string    `json:"job_name"`
	StartDate  time.Time `json:"start_date"`
	ExpiryDate time.Time `json:"expiry_date"`
	BranchIds  []int     `json:"branch_ids" pg:",array"`
	Assignees  []int     `json:"assignees" pg:",array"`
}

type RemoveJobParam struct {
	Id int `json:"id" valid:"required"`
}

type GetCvsParam struct {
	RecruitmentId int    `json:"recruitment_id" valid:"required"`
	NameApplicant string `json:"name_applicant"`
	MediaIDOther     string `json:"media_id_other"`
	MediaID       int    `json:"media_id"`
	DateReceiptCv string `json:"date_receipt_cv"`
	Status        int    `json:"status"`
	CurrentPage   int    `json:"current_page" valid:"required"`
	RowPerPage    int    `json:"row_per_page" valid:"required"`
}

type RemoveCvParam struct {
	Id int `json:"id" valid:"required"`
}

type CreateCvComment struct {
	RecruitmentId int    `json:"recruitment_id" valid:"required"`
	CvId          int    `json:"cv_id" valid:"required"`
	Comment       string `json:"comment" valid:"required"`
	Receiver      []int  `json:"receiver"`
}

type EditCvComment struct {
	Id      int    `json:"id" valid:"required"`
	Comment string `json:"comment" valid:"required"`
}

type GetCvCommentsParam struct {
	RecruitmentId int `json:"recruitment_id" valid:"required"`
	CvId          int `json:"cv_id" valid:"required"`
}

type GetLogCvStatusParam struct {
	CvId int `json:"cv_id" valid:"required"`
}

type RemoveCvCommentParam struct {
	RecruitmentId int `json:"recruitment_id" valid:"required"`
	Id            int `json:"id" valid:"required"`
}

type NumberCvEachStatus struct {
	CvStatus int `json:"cv_status"`
	Amount   int `json:"amount"`
}

type GetCvRecords struct {
	Id            int       `json:"id"`
	FullName      string    `json:"full_name"`
	DateReceiptCv time.Time `json:"date_receipt_cv"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	MediaIDOther       string       `json:"media_id_other"`
	MediaID       int       `json:"media_id"`
	StatusCV      int       `json:"status_cv"`
	FileName      string    `json:"file_name"`
}

type DetailJobParams struct {
	RecruitmentID int    `json:"recruitment_id"`
	FileName      string `json:"file_name"`
	FileContent   string `json:"file_content"`
}

type GetDetailJobFile struct {
	RecruitmentId int `json:"recruitment_id" valid:"required"`
}
type RemoveDetailJobFile struct {
	RecruitmentId int `json:"recruitment_id" valid:"required"`
}
type CreateLogCvStatus struct {
	CvId      int    `json:"cv_id" valid:"required"`
	Status    int    `json:"status" valid:"required"`
	RecruitmentId int       `json:"recruitment_id"`
	Assignees     []int     `json:"assignees"`
	UpdateDay string `json:"update_day" valid:"required"`
}

type InsertCvCommentReturn struct {
	NotiRequestId int
	Assignees     []int
	Body          string
	Link          string
}
