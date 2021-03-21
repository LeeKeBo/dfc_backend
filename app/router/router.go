package router

import (
	"dfc_backend/app/controller"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"time"
)

func initRouter() *gin.Engine {
	r := gin.Default()
	fmt.Println("log/gin_log_" + time.Now().Format("2006-01-02_15_04_05") + ".log")
	f, err := os.Create("log/gin_log_" + time.Now().Format("2006-01-02_15_04_05") + ".log")
	if err != nil {
		fmt.Println(err.Error())
	}
	gin.DefaultWriter = io.MultiWriter(f)
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())
	r.Use(controller.Cors())
	gin.DefaultWriter = f
	apiRouter(r)
	return r
}

func apiRouter(router *gin.Engine) {
	genCode := router.Group("dfcCode")
	{
		genCode.POST("generateCode", controller.HandleGenerateCode)
		genCode.GET("/", controller.GetTest)
		genCode.POST("storeFile", controller.HandleSaveFile)
		genCode.POST("uploadFile", controller.HandleSendFile)
		genCode.POST("downloadFile", controller.HandleDownloadFile)
	}
}

func Run() {
	r := initRouter()
	r.Run(":8000")
}
