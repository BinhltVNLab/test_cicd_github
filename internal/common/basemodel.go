package common

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9/orm"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type BaseModel struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `pg:",soft_delete"`
}

var _ orm.BeforeInsertHook = (*BaseModel)(nil)
var _ orm.BeforeUpdateHook = (*BaseModel)(nil)

func (m *BaseModel) BeforeInsert(ctx context.Context) (context.Context, error) {
	now := utils.TimeNowUTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	return ctx, nil
}

func (m *BaseModel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt = utils.TimeNowUTC()
	return ctx, nil
}
