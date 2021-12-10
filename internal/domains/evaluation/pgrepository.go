package evaluation

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// PgEvaluationRepository : struct repository
type PgEvaluationRepository struct {
	cm.AppRepository
}

// NewPgEvaluationRepository : create repository
func NewPgEvaluationRepository(logger echo.Logger) (repo *PgEvaluationRepository) {
	repo = &PgEvaluationRepository{}
	repo.Init(logger)
	return
}

// InsertEvaluation : insert new target evaluation form
// Params           : createEvaluationParam
// Returns          : evaluation object, error
func (repo *PgEvaluationRepository) InsertEvaluation(createEvaluationParam *param.CreateEvaluationParams) (m.TargetEvaluation, error) {
	evaluationForm := m.TargetEvaluation{
		UserID:         createEvaluationParam.UserID,
		Content:        createEvaluationParam.Content,
		OrganizationID: createEvaluationParam.OrganizationID,
		Status:         createEvaluationParam.Status,
		Quarter:        createEvaluationParam.Quarter,
		Year:           createEvaluationParam.Year,
		UpdatedBy:      createEvaluationParam.UpdatedBy,
	}

	err := repo.DB.Insert(&evaluationForm)

	if err != nil {
		repo.Logger.Error(err)
	}

	return evaluationForm, err
}

// GetEvaluation : get target evaluation form
// Params           : getEvaluationParams
// Returns          : evaluation object, error
func (repo *PgEvaluationRepository) GetEvaluation(evalFormID int, orgID int) (m.TargetEvaluation, error) {
	evalObj := m.TargetEvaluation{}
	err := repo.DB.Model(&evalObj).
		Where("id = ?", evalFormID).
		Where("organization_id = ?", orgID).
		Limit(1).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return evalObj, err
}

func (repo *PgEvaluationRepository) GetEvaluationToExport(evalFormID int, orgID int) (param.ExportExcelRecord, error) {
	var exportExcelRecord param.ExportExcelRecord
	err := repo.DB.Model(&m.TargetEvaluation{}).
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		ColumnExpr("up.branch").
		ColumnExpr("usrs.name as role").
		Column("tge.content", "tge.organization_id", "tge.quarter", "tge.year").
		Join("JOIN user_profiles AS up ON up.user_id = tge.user_id").
		Join("JOIN users AS urs ON urs.id = tge.user_id").
		Join("JOIN user_roles AS usrs ON usrs.id = urs.role_id").
		Where("tge.id = ?", evalFormID).
		Where("tge.organization_id = ?", orgID).
		Limit(1).
		Select(&exportExcelRecord)

	if err != nil {
		repo.Logger.Error(err)
	}

	return exportExcelRecord, err
}

// UpdateEvaluation : update target evaluation form
// Params           : updateEvaluationParams
// Returns          : evaluation object, error
func (repo *PgEvaluationRepository) UpdateEvaluation(orgID int, updateEvaluationParams *param.UpdateEvaluationParams) error {
	evalObj := &m.TargetEvaluation{
		Content:   updateEvaluationParams.Content,
		Status:    updateEvaluationParams.Status,
		Quarter:   updateEvaluationParams.Quarter,
		Year:      updateEvaluationParams.Year,
		UpdatedBy: updateEvaluationParams.UpdatedBy,
	}
	_, err := repo.DB.Model(evalObj).
		Column("content", "status", "quarter", "year", "updated_by", "updated_at").
		Where("id = ?", updateEvaluationParams.EvalFormID).
		Where("organization_id = ?", orgID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

// GetEvaluationList : search evaluation list
// Params : search param
// Returns : []TargetEvaluation, totalRow, err
func (repo *PgEvaluationRepository) GetEvaluationList(orgID int, searchEvaluationListParams *param.SearchEvaluationListParams) ([]param.EvaluationListResponse, int, error) {
	targetEvaluation := []m.TargetEvaluation{}
	evaluationListResponse := []param.EvaluationListResponse{}

	queryObj := repo.DB.Model(&targetEvaluation)
	queryObj.ColumnExpr("tge.*")
	queryObj.ColumnExpr("pro.branch")
	queryObj.ColumnExpr("pro.employee_id")
	queryObj.ColumnExpr("CAST (tge.content->'result' ->'rank' AS INTEGER) AS rank")
	queryObj.ColumnExpr("CAST (tge.content->'result' ->'points' AS REAL) AS point")
	queryObj.ColumnExpr("CONCAT(pro.first_name, ' ', pro.last_name) AS name")
	queryObj.ColumnExpr("tge.updated_by as updated_by_name")
	queryObj.ColumnExpr("pro.avatar as avatar")
	queryObj.Join("join users as usr on usr.id = tge.user_id")
	queryObj.Join("join user_profiles as pro on pro.user_id = usr.id")
	if searchEvaluationListParams.ProjectId != 0 {
		queryObj.Join(", json_array_elements((tge.content->'projects')::json) obj")
		queryObj.Where("obj ->> 'id' = '?'", searchEvaluationListParams.ProjectId)
	}
	queryObj.Where("tge.organization_id = ?", orgID)

	if len(searchEvaluationListParams.UserIds) == 1 {
		queryObj.Where("tge.user_id = ?", searchEvaluationListParams.UserIds[0])
	} else if len(searchEvaluationListParams.UserIds) > 1 {
		queryObj.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, userId := range searchEvaluationListParams.UserIds {
				q = q.WhereOr("tge.user_id = ?", userId)
			}
			return q, nil
		})
	}

	if searchEvaluationListParams.Name != "" {
		profileName := "%" + searchEvaluationListParams.Name + "%"
		queryObj.Where("vietnamese_unaccent(LOWER(pro.first_name)) || ' ' || vietnamese_unaccent(LOWER(pro.last_name)) LIKE vietnamese_unaccent(LOWER(?0))",
			profileName)
	}
	if searchEvaluationListParams.Quarter != 0 {
		queryObj.Where("quarter = ?", searchEvaluationListParams.Quarter)
	}
	if searchEvaluationListParams.Year != 0 {
		queryObj.Where("year = ?", searchEvaluationListParams.Year)
	}
	if searchEvaluationListParams.Status != 0 {
		queryObj.Where("status = ?", searchEvaluationListParams.Status)
	}
	if searchEvaluationListParams.Rank != 0 {
		queryObj.Where("CAST (tge.content->'result' ->'rank' AS INTEGER) = ? ", searchEvaluationListParams.Rank)
	}
	if searchEvaluationListParams.Branch != 0 {
		queryObj.Where("pro.branch = ?", searchEvaluationListParams.Branch)
	}

	queryObj.Offset((searchEvaluationListParams.CurrentPage - 1) * searchEvaluationListParams.RowPerPage)
	queryObj.Order("tge.created_at DESC")
	queryObj.Limit(searchEvaluationListParams.RowPerPage)
	totalRow, err := queryObj.SelectAndCount(&evaluationListResponse)
	return evaluationListResponse, totalRow, err
}

// GetEvaluationListByUserID : get evaluation list by userID
// Params : search param
// Returns : []TargetEvaluation, totalRow, err
func (repo *PgEvaluationRepository) GetEvaluationListByUserID(orgID int, userID int) []param.EvaluationListResponse {
	targetEvaluations := []m.TargetEvaluation{}
	evaluationListResponse := []param.EvaluationListResponse{}

	queryObj := repo.DB.Model(&targetEvaluations)
	queryObj.Where("tge.user_id = ?", userID)
	queryObj.Where("tge.organization_id = ?", orgID)

	_ = queryObj.Select(&evaluationListResponse)

	return evaluationListResponse
}

// DeleteEvaluation : delete evaluation by ID
// Params : evaluationID
// Returns : error
func (repo *PgEvaluationRepository) DeleteEvaluation(evaluationID int) error {
	project := m.TargetEvaluation{}
	_, err := repo.DB.Model(&project).
		Where("id = ?", evaluationID).
		Delete()

	return err
}

// CheckEvaluationExists : check evaluation exist or not by year and quarter
// Params : orgID, userID, year, quarter
// Returns : int, error
func (repo *PgEvaluationRepository) CheckEvaluationExists(orgID, userID, year, quarter int) (m.TargetEvaluation, error) {
	evalObj := m.TargetEvaluation{}
	err := repo.DB.Model(&evalObj).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("year = ?", year).
		Where("quarter = ?", quarter).
		Select()

	return evalObj, err
}

// CheckExist : check evaluation exist or not by year and quarter
// Params : orgID, userID, year, quarter
// Returns : int, error
func (repo *PgEvaluationRepository) CheckExist(orgID int, userID int, year int, quarter int) (int, error) {
	targetEvaluations := []m.TargetEvaluation{}
	count, err := repo.DB.Model(&targetEvaluations).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Where("year = ?", year).
		Where("quarter = ?", quarter).
		Count()

	return count, err
}

func (repo *PgEvaluationRepository) EvaluationRankLastFourQuarter(organizationID int) ([]param.RankLastFourQuarter, error) {
	var records []param.RankLastFourQuarter

	query := repo.DB.Model(&m.TargetEvaluation{})
	query.ColumnExpr("tge.content -> 'result' ->> 'rank' AS rank, tge.quarter, tge.year, COUNT(tge.content -> 'result' ->> 'rank') AS amount")
	query.Where("tge.organization_id = ?", organizationID)

	quarter, year := utils.GetCurrentQuarterAndYear()
	switch quarter {
	case 1:
		query.Where("(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?)", quarter+3, year-1, quarter+2, year-1, quarter+1, year-1, quarter, year-1)
	case 2:
		query.Where("(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?)", quarter-1, year, quarter+2, year-1, quarter+1, year-1, quarter, year-1)
	case 3:
		query.Where("(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?)", quarter-1, year, quarter-2, year, quarter+1, year-1, quarter, year-1)
	case 4:
		query.Where("(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?) OR"+
			"(tge.quarter = ? AND tge.year = ?)", quarter-1, year, quarter-2, year, quarter-3, year, quarter, year-1)
	}
	query.Group("tge.quarter", "tge.year", "rank")
	query.OrderExpr("tge.quarter DESC, tge.year ASC")

	err := query.Select(&records)
	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgEvaluationRepository) SelectCommentTwoConsecutiveQuarter(
	organizationID int,
	params *param.GetCommentTwoConsecutiveQuarterParams,
) ([]param.CommentTwoConsecutiveQuarterRecords, error) {
	var records []param.CommentTwoConsecutiveQuarterRecords
	err := repo.DB.Model(&m.TargetEvaluation{}).
		ColumnExpr("up.user_id AS user_id").
		ColumnExpr("up.first_name || ' ' || up.last_name full_name").
		ColumnExpr("up.avatar AS avatar").
		ColumnExpr("MAX((tge.content -> 'result' ->> 'rank')::int) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS last_rank", params.LastYear, params.LastQuarter).
		ColumnExpr("MAX((tge.content -> 'result' ->> 'rank')::int) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS rank", params.Year, params.Quarter).
		ColumnExpr("MAX((tge.content -> 'result' ->> 'points')::real) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS last_score", params.LastYear, params.LastQuarter).
		ColumnExpr("MAX((tge.content -> 'result' ->> 'points')::real) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS score", params.Year, params.Quarter).
		ColumnExpr("MAX((tge.content -> 'comment' ->> 'superior_cmt')::text) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS last_comment", params.LastYear, params.LastQuarter).
		ColumnExpr("MAX((tge.content -> 'comment' ->> 'superior_cmt')::text) FILTER (WHERE (tge.year, tge.quarter) = (?, ?)) AS comment", params.Year, params.Quarter).
		Join("JOIN user_profiles AS up ON up.user_id = tge.user_id").
		Where("tge.organization_id = ?", organizationID).
		Where("(tge.year, tge.quarter) = (?, ?)", params.LastYear, params.LastQuarter).
		WhereOr("(tge.year, tge.quarter) = (?, ?)", params.Year, params.Quarter).
		Group("up.user_id", "full_name", "up.avatar").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}
