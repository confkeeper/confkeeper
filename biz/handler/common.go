package handler

type Code int32

const (
	Code_Success       Code = 200
	Code_Unauthorized  Code = 401
	Code_Err           Code = 500
	Code_DBErr         Code = 501
	Code_PasswordErr   Code = 502
	Code_AlreadyExists Code = 503
	Code_CaptchaErr    Code = 504
	Code_UserErr       Code = 505
)

type CommonResp struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
}
