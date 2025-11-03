package utils

import "github.com/gin-gonic/gin"

var errCode = map[string]string{
	"1000": "email has exists",
	"1001": "email not exists",
	"1002": "pwd err",
	"1003": "volume is not enough",
	"1004": "need other chuncks",
	"1005": "sec tran success",
	"1006": "chunck has uploaded",
	"1007": "status is finish/cancel",
	"1008": "status is not upload",
	"1009": "missing index",
	"1010": "hash err,file is broken", //next : 重新上传

	"1011": "no such file",
	"1012": "range format err, should [bytes=?-?]",
	"1013": "range over",
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

func ResponseWithCodeAndData(code string, data any, ctx *gin.Context) {
	ctx.JSON(200, Response{Code: code, Msg: errCode[code], Data: data})
}
func ResponseWith400Msg(msg string, ctx *gin.Context) {
	ctx.JSON(400, Response{Code: "1", Msg: msg, Data: nil})
}
func ResponseWith400Code(code string, ctx *gin.Context) {
	ctx.JSON(400, Response{Code: code, Msg: errCode[code], Data: nil})
}
