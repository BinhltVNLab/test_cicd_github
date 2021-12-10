package fcmtoken

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
)

type PgFcmTokenRepository struct {
	cm.AppRepository
}

func NewPgFcmTokenRepository(logger echo.Logger) (repo *PgFcmTokenRepository) {
	repo = &PgFcmTokenRepository{}
	repo.Init(logger)
	return
}

func (repo *PgFcmTokenRepository) InsertFcmToken(params *param.CreateFcmTokenParam) error {
	fcmToken := m.FcmToken{UserId: params.UserId, Token: params.Token}
	err := repo.DB.Insert(&fcmToken)
	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}

func (repo *PgFcmTokenRepository) CountFcmToken(userId int, token string) (int, error) {
	count, err := repo.DB.Model(&m.FcmToken{}).
		Where("user_id = ?", userId).
		Where("token = ?", token).
		Count()

	if err != nil {
		repo.Logger.Error(err)
	}

	return count, err
}

func (repo *PgFcmTokenRepository) SelectFcmTokenByUserId(userId int) ([]string, error) {
	var tokens []string
	err := repo.DB.Model(&m.FcmToken{}).
		Column("token").
		Where("user_id = ?", userId).
		Select(&tokens)

	if err != nil {
		repo.Logger.Error(err)
	}

	return tokens, err
}

func (repo *PgFcmTokenRepository) SelectMultiFcmTokens(usersId []int, currentUserId int) ([]string, error) {
	var records []string
	if len(usersId) == 0 || (len(usersId) == 1 && usersId[0] == currentUserId) {
		return records, nil
	}

	err := repo.DB.Model(&m.FcmToken{}).
		ColumnExpr(" DISTINCT token").
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, userId := range usersId {
				if currentUserId == userId {
					continue
				}
				q = q.WhereOr("user_id = ?", userId)
			}
			return q, nil
		}).
		Select(&records)

	if err != nil {
		repo.Logger.Error(err)
	}

	return records, err
}

func (repo *PgFcmTokenRepository) DeleteFcmToken(token string) error {
	_, err := repo.DB.Model(&m.FcmToken{}).
		Where("token = ?", token).
		Delete()

	if err != nil {
		repo.Logger.Error(err)
	}

	return err
}
