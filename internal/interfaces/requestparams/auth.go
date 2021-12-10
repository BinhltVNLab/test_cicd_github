package requestparams

// CreateRegRequestParams struct for receive param from frontend
type CreateRegRequestParams struct {
	RequestEmail string `json:"email" form:"email" validate:"required,email"`
	GoogleID     string
}

// CheckRegistrationCodeParams struct for receive param from frontend
type CheckRegistrationCodeParams struct {
	Code string `json:"registrationCode" form:"registrationCode" validate:"required"`
}

// GoogleProfile struct for receive param from google api
type GoogleProfile struct {
	GoogleID      string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

// UpdateGoogleUser struct to update google id when login
type UpdateGoogleUser struct {
	UserID   int
	GoogleID string
}

// RegisterUserParams struct to register user
type RegisterUserParams struct {
	OrganizationID int
	Email          string
	Password       string
	RoleID         int
	GoogleID       string
	LanguageId     int
}
