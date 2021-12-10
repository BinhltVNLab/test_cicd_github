package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

// EvaluationRepository interface
type EvaluationRepository interface {
	InsertEvaluation(createEvaluationParams *param.CreateEvaluationParams) (m.TargetEvaluation, error)
	GetEvaluation(evalFormID int, orgID int) (m.TargetEvaluation, error)
	UpdateEvaluation(orgID int, updateEvaluationParams *param.UpdateEvaluationParams) error
	GetEvaluationList(orgID int, searchEvaluationListParams *param.SearchEvaluationListParams) ([]param.EvaluationListResponse, int, error)
	GetEvaluationListByUserID(orgID int, userID int) []param.EvaluationListResponse
	DeleteEvaluation(evaluationID int) error
	CheckEvaluationExists(orgID, userID, year, quarter int) (m.TargetEvaluation, error)
	EvaluationRankLastFourQuarter(organizationID int) ([]param.RankLastFourQuarter, error)
	CheckExist(orgID int, userID int, year int, quarter int) (int, error)
	GetEvaluationToExport(evalFormID int, orgID int) (param.ExportExcelRecord, error)
	SelectCommentTwoConsecutiveQuarter(
		organizationID int,
		params *param.GetCommentTwoConsecutiveQuarterParams,
	) ([]param.CommentTwoConsecutiveQuarterRecords, error)
}
