package common

const (
	RespCodeSuccess    = 0
	RespCodeUnknownErr = -1
)

var defaultMessage = map[int]string{
	RespCodeSuccess:    "Success",
	RespCodeUnknownErr: "Unknown error",
}

type BaseResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func MakeResp(code int, data interface{}) BaseResp {
	return BaseResp{
		Code:    code,
		Message: defaultMessage[code],
		Data:    data,
	}
}

func MakeSuccessResp(data interface{}) BaseResp {
	return MakeResp(RespCodeSuccess, data)
}

func MakeUnknownErrorResp() BaseResp {
	return MakeResp(RespCodeUnknownErr, nil)
}
