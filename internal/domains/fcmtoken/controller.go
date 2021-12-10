package fcmtoken

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	cm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/common"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	"net/http"
)

type Controller struct {
	cm.BaseController

	FcmTokenRepo rp.FcmTokenRepository
}

func NewFcmTokenController(logger echo.Logger, fcmTokenRepo rp.FcmTokenRepository) (ctr *Controller) {
	ctr = &Controller{cm.BaseController{}, fcmTokenRepo}
	ctr.Init(logger)
	return
}

func (ctr *Controller) CreateFcmToken(c echo.Context) error {
	params := new(param.CreateFcmTokenParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	count, err := ctr.FcmTokenRepo.CountFcmToken(params.UserId, params.Token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	if count > 0 {
		return c.JSON(http.StatusMethodNotAllowed, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Token of user already exist",
		})
	}

	if err := ctr.FcmTokenRepo.InsertFcmToken(params); err != nil {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Create token successful",
	})
}

func (ctr *Controller) GetFcmTokens(c echo.Context) error {
	params := new(param.GetFcmTokenParam)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid Params",
			Data:    err,
		})
	}

	if _, err := valid.ValidateStruct(params); err != nil {
		return c.JSON(http.StatusBadRequest, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "Invalid field value",
		})
	}

	tokens, err := ctr.FcmTokenRepo.SelectFcmTokenByUserId(params.UserId)
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		return c.JSON(http.StatusInternalServerError, cf.JsonResponse{
			Status:  cf.FailResponseCode,
			Message: "System Error",
		})
	}

	return c.JSON(http.StatusOK, cf.JsonResponse{
		Status:  cf.SuccessResponseCode,
		Message: "Get token successful",
		Data:    tokens,
	})
}
