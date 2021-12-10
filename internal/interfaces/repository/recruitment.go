package repository

import (
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type RecruitmentRepository interface {
	InsertJob(
		organizationId int,
		createdBy int,
		params *param.CreateJobParam,
		notificationRepo NotificationRepository,
	) (string, string, error)
	InsertDetailedJobRecruitmentFromModel(detailedRecruitment *m.DetailedJobRecruitment) error
	UpdateDetailedJobRecruitmentFromModel(detailedRecruitment *m.DetailedJobRecruitment) error
	FindDetailedJobRecruitmentById(id int, columns ...string) (*m.DetailedJobRecruitment, error)
	UpdateJob(organizationId int, params *param.EditJobParam) error
	SelectJobs(organizationId int, params *param.GetJobsParam) ([]param.GetJobRecords, int, error)
	SelectJob(id int, columns ...string) (m.Recruitment, error)
	SelectCvs(recruitmentId int, params *param.GetCvsParam) ([]param.GetCvRecords, int, error)
	CountJob(id int) (int, error)
	DeleteJob(id int) error
	CreateCvNoti(
		cv *m.Cv,
		organizationId int,
		createdBy int,
		params *param.CreateCvParam,
		notificationRepo NotificationRepository,
	) (string, string, error)
	InsertCv(
		organizationId int,
		createdBy int,
		params *param.UploadCvParam,
		notificationRepo NotificationRepository,
	) (string, string, error)
	InsertCvFromModel(cv *m.Cv) error
	FindCvById(id int, columns ...string) (*m.Cv, error)
	DeleteCv(id int) error
	UpdateCv(cv *m.Cv) error
	CountCvById(id int) (int, error)
	InsertCvComment(
		organizationId int,
		createdBy int,
		params *param.CreateCvComment,
	) (*param.InsertCvCommentReturn, error)
	UpdateCvComment(id int, comment string) error
	SelectCvCommentById(id int, columns ...string) (m.CvComment, error)
	SelectCvCommentsByCvId(cvId int, columns ...string) ([]m.CvComment, error)
	SelectLogCvStates(cvId int, columns ...string) ([]m.LogCvState, error)
	DeleteCvCommentById(id int) error
	StatisticCvByStatus(organizationID int, id int) ([]param.NumberCvEachStatus, error)
	SelectPermissions(organizationId int, userId int) ([]param.SelectPermissionRecords, error)
	UpdateDetailedJobFileName(detailedRecruitment *m.Recruitment) error
	InsertLogCvStateFromModel(logCvState *m.LogCvState) error
	CreateLogCvStatusModel(
		logCvState *m.LogCvState,
		organizationId int,
		createdBy int,
		params *param.CreateLogCvStatus,
		notificationRepo NotificationRepository,)(string, string, error)
}
