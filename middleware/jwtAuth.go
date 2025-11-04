package middleware

import (
	"cloud_store/utils"
	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("Authorization")
		if token == "" {
			utils.ResponseWithMsg("未携带token", ctx)
			ctx.Abort()
			return
		}
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.ResponseWithMsg("非法token", ctx)
			ctx.Abort()
			return
		}
		//limit
		ok,err:=RedisLimitBucket(claims.UserId)
		if err!=nil{
			utils.ResponseWithMsg("[internal server err]",ctx)
			ctx.Abort()
			return 
		}
		if !ok{
			utils.ResponseWithCode("1015",ctx)
			ctx.Abort()
			return 
		}
		ctx.Set("claims", claims)
	}
}
