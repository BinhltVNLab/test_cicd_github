package requestparams

import (
	"time"

	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// CreateEvaluationParams : create target params
type CreateEvaluationParams struct {
	UserID         int       `json:"user_id"`
	Content        m.Content `json:"content"`
	OrganizationID int       `json:"organization_id"`
	Status         int       `json:"status"`
	Quarter        int       `json:"quarter" valid:"required,range(1|4)"`
	Year           int       `json:"year" valid:"required"`
	UpdatedBy      int       `json:"updated_by"`
}

// DuplicateEvaluationParams : duplicate target params
type DuplicateEvaluationParams struct {
	EvalFormID int `json:"eval_form_id" valid:"required"`
}

// GetEvaluationParams : get evaluation form params
type GetEvaluationParams struct {
	EvalFormID int `json:"eval_form_id" valid:"required"`
}

// UpdateEvaluationParams : update evaluation form param
type UpdateEvaluationParams struct {
	EvalFormID int       `json:"eval_form_id" valid:"required"`
	Content    m.Content `json:"content" valid:"required"`
	Status     int       `json:"status" valid:"required"`
	Quarter    int       `json:"quarter" valid:"required,range(1|4)"`
	Year       int       `json:"year" valid:"required"`
	UpdatedBy  int       `json:"updated_by"`
}

// SearchEvaluationListParams : get evaluation list param
type SearchEvaluationListParams struct {
	UserIds     []int  `json:"user_ids"`
	Name        string `json:"name"`
	ProjectId   int    `json:"project_id"`
	Quarter     int    `json:"quarter"`
	Year        int    `json:"year"`
	Branch      int    `json:"branch"`
	Rank        int    `json:"rank"`
	Status      int    `json:"status"`
	CurrentPage int    `json:"current_page"`
	RowPerPage  int    `json:"row_per_page"`
}

// EvaluationListResponse : struct for target evaluation database table
type EvaluationListResponse struct {
	tableName struct{} `pg:",discard_unknown_columns"`

	ID            int
	Name          string
	UpdatedByName string
	Quarter       int
	Year          int
	Branch        int
	Rank          int
	Status        int
	Point         float64
	EmployeeId    int
	UpdatedAt     time.Time
	UpdatedBy     int
	Avatar        string
}

// DeleteEvaluationParams : get evaluation form params
type DeleteEvaluationParams struct {
	EvaluationID int `json:"evaluation_id" valid:"required"`
}

// CheckEvalExistParams : get evaluation form params
type CheckEvalExistParams struct {
	Quarter int `json:"quarter" valid:"required,range(1|4)"`
	Year    int `json:"year" valid:"required"`
}

// ExportExcelParam : Export excel param
type ExportExcelParam struct {
	ID int `json:"evaluation_id" valid:"required"`
}

type ExportExcelRecord struct {
	FullName       string
	Branch         int
	Role           string
	Content        m.Content
	OrganizationID int
	Quarter        int
	Year           int
}

// RankLastFourQuarter : struct for rank of the last four quarter
type RankLastFourQuarter struct {
	Rank    int `json:"rank"`
	Quarter int `json:"quarter"`
	Year    int `json:"year"`
	Amount  int `json:"amount"`
}

// EvaluationRankDatasets : struct for rank of the last four quarter
type EvaluationRankDatasets struct {
	Datetime []string  `json:"datetime"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset : struct for data set
type Dataset struct {
	Rank string `json:"rank"`
	Data [4]int `json:"data"`
}

type GetCommentTwoConsecutiveQuarterParams struct {
	Quarter     int `json:"quarter" valid:"required,range(1|4)"`
	Year        int `json:"year" valid:"required"`
	LastQuarter int `json:"last_quarter" valid:"required,range(1|4)"`
	LastYear    int `json:"last_year" valid:"required"`
}

type CommentTwoConsecutiveQuarterRecords struct {
	UserID      int     `json:"user_id"`
	FullName    string  `json:"full_name"`
	Avatar      string  `json:"avatar"`
	Rank        int     `json:"rank"`
	Score       float64 `json:"score"`
	Comment     string  `json:"comment"`
	LastRank    int     `json:"last_rank"`
	LastScore   float64 `json:"last_score"`
	LastComment string  `json:"last_comment"`
}
