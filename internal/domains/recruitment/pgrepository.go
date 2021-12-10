package recruitment

import (
	"fmt"
	"strconv"
	"time"

	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils/calendar"
)

type PgRecruitmentRepository struct {
	cm.AppRepository
}

func NewPgRecruitmentRepository(logger echo.Logger) (repo *PgRecruitmentRepository) {
	repo = &PgRecruitmentRepository{}
	repo.Init(logger)
	return
}

func (repo *PgRecruitmentRepository) InsertJob(
	organizationId int,
	createdBy int,
	params *param.CreateJobParam,
	notificationRepo rp.NotificationRepository,
) (string, string, error) {
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var transErr error
		job := m.Recruitment{
			OrganizationId: organizationId,
			JobName:        params.JobName,
			StartDate:      calendar.ParseTime(cf.FormatDateDatabase, params.StartDate),
			ExpiryDate:     calendar.ParseTime(cf.FormatDateDatabase, params.ExpiryDate),
			BranchIds:      params.BranchIds,
			Assignees:      params.Assignees,
		}

		transErr = tx.Insert(&job)
		if transErr != nil {
			return transErr
		}

		if len(params.Assignees) > 0 {
			notificationParams := new(param.InsertNotificationParam)
			notificationParams.Content = "has added you to a recruiting job"
			notificationParams.RedirectUrl = "/recruitment/recruitment-details?recruitment_id=" + strconv.Itoa(job.ID)

			for _, userId := range params.Assignees {
				if userId == createdBy {
					continue
				}
				notificationParams.Receiver = userId
				transErr = notificationRepo.InsertNotificationWithTx(tx, organizationId, createdBy, notificationParams)
				if transErr != nil {
					return transErr
				}
			}

			body = notificationParams.Content
			link = notificationParams.RedirectUrl
		}

		return transErr
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}

func (repo *PgRecruitmentRepository) InsertDetailedJobRecruitmentFromModel(detailedRecruitment *m.DetailedJobRecruitment) error {
	return repo.DB.Insert(detailedRecruitment)
}

func (repo *PgRecruitmentRepository) UpdateJob(organizationId int, params *param.EditJobParam) error {
	job := m.Recruitment{
		OrganizationId: organizationId,
		JobName:        params.JobName,
		StartDate:      calendar.ParseTime(cf.FormatDateDatabase, params.StartDate),
		BranchIds:      params.BranchIds,
		Assignees:      params.Assignees,
	}

	if params.ExpiryDate != "" {
		job.ExpiryDate = calendar.ParseTime(cf.FormatDateDatabase, params.ExpiryDate)
	}

	_, err := repo.DB.Model(&job).
		Where("id = ?", params.Id).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgRecruitmentRepository) CountJob(id int) (int, error) {
	count, err := repo.DB.Model(&m.Recruitment{}).
		Where("id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgRecruitmentRepository) UpdateDetailedJobRecruitmentFromModel(detailedRecruitment *m.DetailedJobRecruitment) error {
	_, err := repo.DB.Model(detailedRecruitment).
		Where("id = ?", detailedRecruitment.ID).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
		return err
	}

	return nil
}

func (repo *PgRecruitmentRepository) FindDetailedJobRecruitmentById(id int, columns ...string) (*m.DetailedJobRecruitment, error) {
	detailedJob := m.DetailedJobRecruitment{}
	query := repo.DB.Model(&detailedJob).Where("id = ?", id).Limit(1)
	if len(columns) > 0 {
		query = query.Column(columns...)
	}
	if err := query.Select(); err != nil {
		repo.Logger.Error(err)
		return nil, err
	}

	return &detailedJob, nil
}

func (repo *PgRecruitmentRepository) SelectJobs(organizationId int, params *param.GetJobsParam) ([]param.GetJobRecords, int, error) {
	var records []param.GetJobRecords
	q := repo.DB.Model(&m.Recruitment{}).
		Column("id", "job_name", "start_date", "expiry_date", "branch_ids", "assignees").
		Where("organization_id = ?", organizationId)

	now := time.Now().Format(cf.FormatDateDatabase)
	if params.JobName != "" {
		q.Where("vietnamese_unaccent(LOWER(job_name)) LIKE vietnamese_unaccent(LOWER(?))", "%"+params.JobName+"%")
	}

	if params.JobStatus == cf.JOBACTIVE {
		q.Where("DATE(start_date) <= to_date(?,'YYYY-MM-DD')", string(now))
		q.Where("DATE(expiry_date) >= to_date(?,'YYYY-MM-DD')", string(now))
	} else if params.JobStatus == cf.JOBDONE {
		q.Where("DATE(expiry_date) < to_date(?,'YYYY-MM-DD')", string(now))
	}

	if params.ExpiryDate != "" {
		q.Where("DATE(expiry_date) <= to_date(?,'YYYY-MM-DD')", params.ExpiryDate)
	}

	if params.BranchId != 0 {
		q.Where("? = ANY (branch_ids)", params.BranchId)
	}

	if params.UserID != 0 {
		q.Where("? = ANY (assignees)", params.UserID)
	}

	totalRow, err := q.Offset((params.CurrentPage - 1) * params.RowPerPage).
		Limit(params.RowPerPage).
		Order("created_at DESC").
		SelectAndCount(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, totalRow, err
}

func (repo *PgRecruitmentRepository) SelectJob(id int, columns ...string) (m.Recruitment, error) {
	var recruitment m.Recruitment
	err := repo.DB.Model(&recruitment).
		Column(columns...).
		Where("id = ?", id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return recruitment, err
}

// Note: DeleteJob will delete DetailedJobRecruitment as well
func (repo *PgRecruitmentRepository) DeleteJob(id int) error {
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.Model(&m.Recruitment{}).
			Where("id = ?", id).
			Delete()

		_, err = tx.Model(&m.DetailedJobRecruitment{}).
			Where("recruitment_id = ?", id).
			Delete()

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgRecruitmentRepository) InsertCvFromModel(cv *m.Cv) error {
	return repo.DB.Insert(cv)
}
func (repo *PgRecruitmentRepository) CreateCvNoti(
	cv *m.Cv,
	organizationId int,
	createdBy int,
	params *param.CreateCvParam,
	notificationRepo rp.NotificationRepository,
) (string, string, error) {

	var body string
	var link string
	t, _ := time.Parse(cf.FormatDateDatabase, params.DateReceiptCv)
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		cv := m.Cv{
			BaseModel: cm.BaseModel{
				ID: cv.ID,
			},
			RecruitmentId:   params.RecruitmentId,
			MediaId:         params.MediaId,
			FileName:        params.FileName,
			FullName:        params.FullName,
			PhoneNumber:     params.PhoneNumber,
			DateReceiptCv:   t,
			Email:           params.Email,
			InterviewMethod: params.InterviewMethod,
			Salary:          params.Salary,
			ContactLink:     params.ContactLink,
		}
		if err := tx.Insert(&cv); err != nil {
			return err
		}
		if len(params.Assignees) > 0 {
			notificationParams := new(param.InsertNotificationParam)
			notificationParams.Content = "has added new cv"
			notificationParams.RedirectUrl = "/recruitment/recruitment-details?recruitment_id=" + strconv.Itoa(params.RecruitmentId) +  
			"&cv_id=" + strconv.Itoa(cv.ID)

			for _, userId := range params.Assignees {
				if userId == createdBy {
					continue
				}
				notificationParams.Receiver = userId
				fmt.Println("rtfytytytytytytytytytyty")
				errTx = notificationRepo.InsertNotificationWithTx(tx, organizationId, createdBy, notificationParams)
				if errTx != nil {
					return errTx
				}
			}

			body = notificationParams.Content
			link = notificationParams.RedirectUrl
		}

		return errTx
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}

func (repo *PgRecruitmentRepository) FindCvById(id int, columns ...string) (*m.Cv, error) {
	cv := m.Cv{}
	query := repo.DB.Model(&cv).Where("id = ?", id).Limit(1)
	if len(columns) > 0 {
		query = query.Column(columns...)
	}
	if err := query.Select(); err != nil {
		repo.Logger.Error(err)
		return nil, err
	}

	return &cv, nil
}

func (repo *PgRecruitmentRepository) InsertCv(
	organizationId int,
	createdBy int,
	params *param.UploadCvParam,
	notificationRepo rp.NotificationRepository,
) (string, string, error) {
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		var errTx error
		for _, cvField := range params.CvFields {
			cv := m.Cv{
				RecruitmentId: params.RecruitmentId,
				MediaId:       cvField.MediaId,
				FileName:      cvField.FileName,
			}

			errTx := tx.Insert(&cv)
			if errTx != nil {
				repo.Logger.Error(errTx)
				return errTx
			}
		}

		if len(params.Assignees) > 0 {
			notificationParams := new(param.InsertNotificationParam)
			notificationParams.Content = "has added new cv"
			notificationParams.RedirectUrl = "/recruitment/recruitment-details?recruitment_id=" + strconv.Itoa(params.RecruitmentId)

			for _, userId := range params.Assignees {
				if userId == createdBy {
					continue
				}
				notificationParams.Receiver = userId
				errTx = notificationRepo.InsertNotificationWithTx(tx, organizationId, createdBy, notificationParams)
				if errTx != nil {
					return errTx
				}
			}

			body = notificationParams.Content
			link = notificationParams.RedirectUrl
		}

		return errTx
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}

func (repo *PgRecruitmentRepository) SelectCvs(recruitmentId int, params *param.GetCvsParam) ([]param.GetCvRecords, int, error) {
	var cvs []param.GetCvRecords
	q := repo.DB.Model(&m.Cv{})
	var totalRow int

	q.Column("c.id", "c.full_name", "c.date_receipt_cv", "c.file_name", "c.media_id", "c.media_id_other", "c.status_cv").
		ColumnExpr("c.updated_at as last_updated_at").
		Where("c.recruitment_id = ?", recruitmentId)

	if params.DateReceiptCv != "" {
		q.Where("DATE(c.date_receipt_cv) = to_date(?,'YYYY-MM-DD')", params.DateReceiptCv)
	}

	if params.NameApplicant != "" {
		q.Where("vietnamese_unaccent(LOWER(c.full_name)) LIKE vietnamese_unaccent(LOWER(?0))", "%"+params.NameApplicant+"%")
	}

	if params.MediaID != 0 {
		q.Where("c.media_id = ?", params.MediaID)
	}

	if params.Status != 0 {
		q.Where("c.status_cv = ?", params.Status)
	}

	q.Offset((params.CurrentPage - 1) * params.RowPerPage).
		Order("c.updated_at DESC").
		Limit(params.RowPerPage)

	totalRow, err := q.SelectAndCount(&cvs)

	if err != nil {
		repo.Logger.Errorf("%+v", err)
	}

	return cvs, totalRow, err
}

func (repo *PgRecruitmentRepository) DeleteCv(id int) error {
	_, err := repo.DB.Model(&m.Cv{}).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgRecruitmentRepository) CountCvById(id int) (int, error) {
	count, err := repo.DB.Model(&m.Cv{}).
		Where("id = ?", id).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgRecruitmentRepository) UpdateCv(cv *m.Cv) error {
	_, err := repo.DB.Model(cv).
		Where("id = ?", cv.ID).
		UpdateNotZero()

	if err != nil {
		repo.Logger.Error(err)
		return err
	}

	return nil
}

func (repo *PgRecruitmentRepository) InsertCvComment(
	organizationId int,
	createdBy int,
	params *param.CreateCvComment,
) (*param.InsertCvCommentReturn, error) {
	var (
		notiRequestId int
		assignees     = make([]int, 0, len(params.Receiver))
		body          string
		link          string
	)

	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
		cvComment := m.CvComment{
			CvId:      params.CvId,
			CreatedBy: createdBy,
			Comment:   params.Comment,
			Receivers: params.Receiver,
		}
		if err := tx.Insert(&cvComment); err != nil {
			return err
		}

		notiRequest := m.NotiRequest{
			Status: m.NotiRequestStatusInitial,
		}
		if err := tx.Insert(&notiRequest); err != nil {
			return err
		}
		notiRequestId = notiRequest.ID
		body = "added a comment"
		link = "/recruitment/recruitment-details?recruitment_id=" + strconv.Itoa(params.RecruitmentId) +
			"&cv_id=" + strconv.Itoa(cvComment.CvId) + "&comment_id=" + strconv.Itoa(cvComment.ID)

		for _, receiverId := range params.Receiver {
			if receiverId == createdBy || utils.FindIntInSlice(assignees, receiverId) {
				continue
			}

			if err := tx.Insert(&m.Notification{
				OrganizationId: organizationId,
				Sender:         createdBy,
				Receiver:       receiverId,
				Title:          "Micro Erp New Notification",
				Content:        body,
				Status:         cf.NotificationStatusUnread,
				RedirectUrl:    link,
				NotiRequestId:  notiRequestId,
			}); err != nil {
				return err
			}

			assignees = append(assignees, receiverId)
		}

		return nil
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return &param.InsertCvCommentReturn{
		NotiRequestId: notiRequestId,
		Assignees:     assignees,
		Body:          body,
		Link:          link,
	}, err
}

func (repo *PgRecruitmentRepository) UpdateCvComment(id int, comment string) error {
	cvComment := m.CvComment{Comment: comment}
	_, err := repo.DB.Model(&cvComment).Where("id = ?", id).UpdateNotZero()
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgRecruitmentRepository) SelectCvCommentById(id int, columns ...string) (m.CvComment, error) {
	var cvComment m.CvComment
	err := repo.DB.Model(&cvComment).
		Column(columns...).
		Where("id = ?", id).
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return cvComment, err
}

func (repo *PgRecruitmentRepository) SelectLogCvStates(cvId int, columns ...string) ([]m.LogCvState, error) {
	var logCvsStatus []m.LogCvState
	err := repo.DB.Model(&logCvsStatus).
		Column(columns...).
		Where("cv_id = ?", cvId).
		Select(&logCvsStatus)

	if err != nil {
		repo.Logger.Error(err)
	}

	return logCvsStatus, err
}

func (repo *PgRecruitmentRepository) SelectCvCommentsByCvId(cvId int, columns ...string) ([]m.CvComment, error) {
	var cvComments []m.CvComment
	err := repo.DB.Model(&cvComments).
		Column(columns...).
		Where("cv_id = ?", cvId).
		Order("created_at ASC").
		Select()

	if err != nil {
		repo.Logger.Error(err)
	}

	return cvComments, err
}

func (repo *PgRecruitmentRepository) DeleteCvCommentById(id int) error {
	_, err := repo.DB.Model(&m.CvComment{}).
		Where("id = ?", id).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgRecruitmentRepository) StatisticCvByStatus(organizationID int, recruitmentId int) ([]param.NumberCvEachStatus, error) {
	var records []param.NumberCvEachStatus
	err := repo.DB.Model(&m.Recruitment{}).
		ColumnExpr("c.status AS cv_status").
		ColumnExpr("COUNT(c.status) AS amount").
		Join("JOIN cvs AS c ON c.recruitment_id = recruitment.id").
		Where("c.deleted_at IS NULL").
		Where("recruitment.id = ?", recruitmentId).
		Where("recruitment.organization_id = ?", organizationID).
		Group("c.status").
		Order("c.status ASC").
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgRecruitmentRepository) SelectPermissions(organizationId int, userId int) ([]param.SelectPermissionRecords, error) {
	var records []param.SelectPermissionRecords
	err := repo.DB.Model(&m.UserPermission{}).
		Column("id", "user_id", "modules").
		Where("up.organization_id = ? AND up.user_id = ?", organizationId, userId).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgRecruitmentRepository) UpdateDetailedJobFileName(detailedRecruitment *m.Recruitment) error {
	recObj := &m.Recruitment{
		DetailJobFileName: detailedRecruitment.DetailJobFileName,
	}
	_, err := repo.DB.Model(recObj).
		Column("detail_job_file_name").
		Where("id = ?", detailedRecruitment.ID).
		Update()

	if err != nil {
		repo.Logger.Error(err)
	}

	return nil
}

func (repo *PgRecruitmentRepository) InsertLogCvStateFromModel(logCvState *m.LogCvState) error {
	err := repo.DB.Insert(logCvState)

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
func (repo *PgRecruitmentRepository) CreateLogCvStatusModel(
	logCvState *m.LogCvState,
	organizationId int,
	createdBy int,
	params *param.CreateLogCvStatus,
	notificationRepo rp.NotificationRepository,
) (string, string, error) {
	var body string
	var link string
	err := repo.DB.RunInTransaction(func(tx *pg.Tx) error {
	var errTx error
	t, _ := time.Parse(cf.FormatDateNoSec, params.UpdateDay)
		logCvState := m.LogCvState{
			CvId:      params.CvId,
			Status:    params.Status,
			UpdateDay: t,
		}
		if err := tx.Insert(&logCvState); err != nil {
			return err
		}
		if len(params.Assignees) > 0 {
			notificationParams := new(param.InsertNotificationParam)
			notificationParams.Content = "has update cv status"
			notificationParams.RedirectUrl = "/recruitment/recruitment-details?recruitment_id=" + strconv.Itoa(params.RecruitmentId) +
			"&cv_id=" + strconv.Itoa(logCvState.CvId)

			for _, userId := range params.Assignees {
				if userId == createdBy {
					continue
				}
				notificationParams.Receiver = userId
				errTx = notificationRepo.InsertNotificationWithTx(tx, organizationId, createdBy, notificationParams)
				if errTx != nil {
					return errTx
				}
			}

			body = notificationParams.Content
			link = notificationParams.RedirectUrl
		}

		return errTx
	})

	if err != nil {
		repo.Logger.Error(err)
	}

	return body, link, err
}
