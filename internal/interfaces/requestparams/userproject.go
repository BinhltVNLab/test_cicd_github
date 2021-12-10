package requestparams

import "time"

// InsertUserProjectParams: struct for insert user project
type InsertUserProjectParams struct {
	UserID    int `json:"user_id" valid:"required"`
	ProjectID int `json:"project_id" valid:"required"`
}

// NumberPeopleProject : number people project
type NumberPeopleProject struct {
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
	ManagedBy   int    `json:"managed_by"`
	Amount      int    `json:"amount"`
}

// GetUserProjectParams : get evaluation list param
type GetUserProjectParams struct {
	ProjectID   int `json:"project_id" valid:"required"`
}

// UserForEachProject : Get user for each project param
type UserForEachProject struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Branch    int       `json:"branch"`
}

// RemoveUserFromProjectParam : Remove user from project param
type RemoveUserFromProjectParam struct {
	ID int `json:"id" valid:"required"`
}

type ProjectUserJoinParam struct {
	UserId int `json:"user_id" valid:"required"`
}

type UserProjectInfoRecords struct {
	UserId   int    `json:"user_id"`
	FullName string `json:"full_name"`
}

type ProjectUserJoinRecords struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	JoinedAt  string `json:"joined_at"`
	CreatedAt string `json:"created_at"`
}
