package requestparams

import (
	"time"
)

// RequestRegistrationParams struct for receive param from frontend
type RequestRegistrationParams struct {
	EmailAddr      string `json:"email" form:"email" validate:"required,email"`
	OrganizationID int    `json:"organizationID" form:"organizationID" validate:"required,numberic"`
	Message        string `json:"message" form:"message"`
}

// InviteUserParams struct for receive param from frontend
type InviteUserParams struct {
	EmailAddr []string `json:"emailList" form:"emailList" validate:"required"`
}

// ManageRequestParams struct filter, pagination request data in page manage rquest
type ManageRequestParams struct {
	OrganizationID int       `json:"organization_id"`
	Email          string    `json:"email" valid:"length(3|1000)~Email at least 3 character"`
	TypeRequest    int       `json:"type" valid:"-"`
	StatusRequest  int       `json:"status" valid:"-"`
	DateFrom       time.Time `json:"date_from"`
	DateTo         time.Time `json:"date_to"`
	CurrentPage    int       `json:"current_page" valid:"-"`
	RowPerPage     int       `json:"row_perpage"`
}

// UpdateRequestStatusParams struct update status request acept or deny
type UpdateRequestStatusParams struct {
	RequestID int    `json:"request_id" valid:"required"`
	Email     string `json:"email"`
	Status    int    `json:"status_request" valid:"required"`
}

type ResendEmailParams struct {
	RequestID int    `json:"request_id"`
	Email     string `json:"email"`
}

type DownloadTemplateParam struct {
	TypeFile string `json:"type_file" valid:"required"`
}
