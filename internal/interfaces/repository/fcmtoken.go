package repository

import param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"

type FcmTokenRepository interface {
	InsertFcmToken(params *param.CreateFcmTokenParam) error
	CountFcmToken(userId int, token string) (int, error)
	SelectFcmTokenByUserId(userId int) ([]string, error)
	SelectMultiFcmTokens(usersId []int, currentUserId int) ([]string, error)
	DeleteFcmToken(token string) error
}
