package common

import (
	"autograph-backend-controller/logging"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func LogRequest(ctx *gin.Context) {
	startTime := time.Now()
	ctx.Next()
	endTime := time.Now()

	cost := endTime.Sub(startTime)
	url := ctx.Request.URL
	method := ctx.Request.Method
	code := ctx.Writer.Status()
	fmt.Println(method)
	logging.Default().Infof("[%4s]%s -> [%d] %s",
		method,
		url,
		code,
		cost,
	)
}

const (
	RequestContextKeyUser = "user"
)

type UserInfo struct {
	Email string
}

func RejectNotLogin(debugMode bool) func(ctx *gin.Context) {
	if debugMode {
		return func(ctx *gin.Context) {
			_, exist := ctx.Get(RequestContextKeyUser)
			if !exist {
				logging.Default().Errorf("not login")
				return
			}

			ctx.Next()
		}
	}

	return func(ctx *gin.Context) {
		_, exist := ctx.Get(RequestContextKeyUser)
		if !exist {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Next()
	}
}

func SetUserInfo(debugMode bool) func(ctx *gin.Context) {
	if debugMode {
		return func(ctx *gin.Context) {
			ctx.Set(RequestContextKeyUser, &UserInfo{
				Email: "autograph_receiver@163.com",
			})

			ctx.Next()
		}
	}

	return func(ctx *gin.Context) {
		// TODO
		auth := ctx.GetHeader("Authorization")
		if len(auth) == 0 {

		}

		userInfo := UserInfo{}
		ctx.Set(RequestContextKeyUser, &userInfo)

		ctx.Next()
	}
}
