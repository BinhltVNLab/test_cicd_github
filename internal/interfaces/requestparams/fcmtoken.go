package requestparams

type CreateFcmTokenParam struct {
	UserId int    `json:"user_id" valid:"required"`
	Token  string `json:"token" valid:"required"`
}

type GetFcmTokenParam struct {
	UserId int `json:"user_id" valid:"required"`
}
