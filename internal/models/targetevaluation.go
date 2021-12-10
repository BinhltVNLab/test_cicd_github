package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

// TargetEvaluation : struct for target evaluation database table
type TargetEvaluation struct {
	cm.BaseModel

	tableName      struct{} `sql:"alias:tge"`
	ID             int
	UserID         int
	Content        Content
	OrganizationID int
	Status         int
	Quarter        int
	Year           int
	UpdatedBy      int
}

// Content : jsonb type
type Content struct {
	Common     Common       `json:"common"`
	Individual []Individual `json:"individuals"`
	Projects   []Projects   `json:"projects"`
	Challenges []Challenge  `json:"challenges"`
	Comment    Comment      `json:"comment"`
	Result     Result       `json:"result"`
}

// Common : jsonb type
type Common struct {
	Value          string  `json:"value" valid:"required"`
	Numeric        float64 `json:"numeric" valid:"required"`
	ActualEval     float64 `json:"actual_eval"`
	CompletionRate float64 `json:"completion_rate"`
	Points         float64 `json:"points"`
	Weight         int     `json:"weight" valid:"required"`
}

// Individual : jsonb type
type Individual struct {
	Weight         int     `json:"weight"`
	Item           string  `json:"item"`
	Goal           int     `json:"goal"`
	ActualEval     float64 `json:"actual_eval"`
	CompletionRate float64 `json:"completion_rate"`
	Points         float64 `json:"points"`
}

// Projects : jsonb type
type Projects struct {
	ID             int     `json:"id"`
	SelfAssessment float64 `json:"self_assessment"`
	SuperiorEval   float64 `json:"superior_eval"`
	Points         float64 `json:"points"`
	Weight         int     `json:"weight"`
}

// Challenge : jsonb type
type Challenge struct {
	Name           string  `json:"name" valid:"required"`
	Actions        string  `json:"actions" valid:"required"`
	SelfAssessment float64 `json:"self_assessment"`
	SuperiorEval   float64 `json:"superior_eval"`
	Points         float64 `json:"points"`
	Weight         int     `json:"weight" valid:"required"`
}

// Comment : jsonb type
type Comment struct {
	SelfCmt     string `json:"self_cmt"`
	SuperiorCmt string `json:"superior_cmt"`
	JikenCmt    string `json:"jiken_cmt"`
}

// Result : jsonb type
type Result struct {
	TotalActualEval float64 `json:"total_actual_eval"`
	CompletionRate  float64 `json:"completion_rate"`
	Points          float64 `json:"points"`
	Weight          int     `json:"weight"`
	Rank            int     `json:"rank"`
}
