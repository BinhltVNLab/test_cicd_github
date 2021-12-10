package models

import (
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	"time"
)

type KanbanTask struct {
	cm.BaseModel

	tableName      struct{} `sql:"alias:kt"`
	KanbanListId   int
	Title          string
	Description    string
	DueDate        time.Time
	Assignees      []int `pg:",array"`
	Status         int
	PositionInList int
	CreatedBy      int
	UpdatedBy      int
	Checklists     []Checklist
}

type Checklist struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}
