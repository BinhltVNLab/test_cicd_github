package models

import (
	"time"

	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
)

type AssetLog struct {
	cm.BaseModel
	tableName      struct{} `sql:"alias:asset_logs"`
	OrganizationID int
	AssetId        int
	UserId         int
	StartDayUsing  time.Time
	EndDayUsing    time.Time
}
