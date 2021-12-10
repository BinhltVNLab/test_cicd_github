package configs

type JsonResponse struct {
	Status  int         `json:"status"`  // ex : 0 : fail , 1 : success
	Message string      `json:"message"` // message error
	Data    interface{} `json:"data"`    // data response
}

const (
	FailResponseCode    = 0
	SuccessResponseCode = 1
	WarningResponseCode = 2
)
