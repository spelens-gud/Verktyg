package kserver

import (
	"context"
	"net/http"

	svrlessgin "github.com/Just-maple/serverless-gin"
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/interfaces/ierror"
	"github.com/spelens-gud/Verktyg/kits/kcontext"
	"github.com/spelens-gud/Verktyg/kits/kerror/errorx"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

type (
	// Deprecated: 使用 GinRouterRegister
	GinSvcRegister interface {
		RegisterRouter(router gin.IRouter, svcH svrlessgin.GinSvcHandler)
	}

	DefaultServiceHandler struct {
		DefaultErrorStatus  int
		ErrorStatus         func(err error) int
		ReturnStructBuilder func(data interface{}, code int, message string, err error) interface{}
	}

	DefaultRet struct {
		Code    int         `json:"code"`
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data"`
	}

	Binder interface {
		Bind(c *gin.Context) (err error)
	}
)

func (svcH *DefaultServiceHandler) ParamHandler(c *gin.Context, params []interface{}) (ok bool) {
	for i, param := range params {
		switch p := param.(type) {
		// 自定义绑定方法
		case Binder:
			if err := p.Bind(c); err != nil {
				// 如果没有定义错误类型 默认使用 参数错误
				if _, ok := err.(ierror.CodeError); !ok {
					err = errorx.ErrInvalidArgument("param bind error: %v", err)
				}
				svcH.Response(c, nil, err)
				return
			}
		// 默认绑定方法
		default:
			if err := c.ShouldBind(params[i]); err != nil {
				err = errorx.ErrInvalidArgument("param bind error: %v", err)
				svcH.Response(c, nil, err)
				return
			}
		}
	}
	return true
}

func (svcH *DefaultServiceHandler) Response(c *gin.Context, data interface{}, err error) {
	var res interface{}
	if err == nil {
		if svcH.ReturnStructBuilder != nil {
			res = svcH.ReturnStructBuilder(data, 0, "", nil)
		} else {
			res = &DefaultRet{Data: data}
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var (
		code       int
		statusCode int
		message    = err.Error()
		ctx        = c.Request.Context()
	)

	// 打印错误
	logger.FromContext(ctx).Errorf("service error: %s", message)

	errorStatusMapping := svcH.ErrorStatus
	if errorStatusMapping == nil {
		errorStatusMapping = DefaultErrorStatus
	}

	// 自定义错误映射
	if statusCode = errorStatusMapping(err); statusCode == 0 {
		// 错误类型判断
		switch e := err.(type) {
		// 错误类型映射
		case ierror.CodeError:
			code = int(e.Code())
			statusCode = e.Code().HttpStatus()
		case ierror.HttpStatusError:
			statusCode = e.HttpStatus()
		default:
			statusCode = svcH.DefaultErrorStatus
		}
		// 不能识别的错误默认返回500
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	}

	if svcH.ReturnStructBuilder != nil {
		res = svcH.ReturnStructBuilder(data, code, message, err)
	} else {
		res = &DefaultRet{
			Code:    code,
			Message: message,
		}
	}

	c.JSON(statusCode, res)

	// 保存相关信息到request context
	kcontext.SetRequestError(c.Request, err)
	kcontext.SetRequestServiceCode(c.Request, code)
}

func DefaultErrorStatus(err error) (errorStatus int) {
	switch err {
	case nil:
		return http.StatusOK
	case context.Canceled:
		return 499
	case context.DeadlineExceeded:
		return http.StatusRequestTimeout
	default:
		return 0
	}
}
