package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Code int32

const (
	Code_Success       Code = 200
	Code_ParamErr      Code = 400
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

type CommonJSONResp struct {
	CommonResp
	Data any `json:"data,omitempty"`
}

func JSON(c *gin.Context, httpStatus int, code Code, msg string) {
	c.JSON(httpStatus, &CommonJSONResp{
		CommonResp: CommonResp{
			Code: code,
			Msg:  msg,
		},
	})
}

func JSONData(c *gin.Context, httpStatus int, code Code, msg string, data any) {
	c.JSON(httpStatus, &CommonJSONResp{
		CommonResp: CommonResp{
			Code: code,
			Msg:  msg,
		},
		Data: data,
	})
}

func ParamError(c *gin.Context, err error) {
	JSON(c, http.StatusBadRequest, Code_ParamErr, "参数错误: "+err.Error())
}
