package utils

import "github.com/gin-gonic/gin"

var errCode = map[string]string{
	"1000":"email has exists",
}

type Response struct {
	Code string
	Msg  string
	Data any
}

func ResponseWithData(data any, ctx *gin.Context) {
	ctx.JSON(200, Response{Code: "0", Msg: "success", Data: data})
}

func ResponseWithMsg(msg string, ctx *gin.Context) {
	ctx.JSON(200, Response{Code: "1", Msg: msg, Data: nil})
}

func ResponseWithCode(code string, ctx *gin.Context) {
	ctx.JSON(200, Response{Code: code, Msg: errCode[code], Data: nil})
}
