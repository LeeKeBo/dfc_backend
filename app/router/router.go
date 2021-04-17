package router

import (
	"dfc_backend/app/controller"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
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
	// 三个接口：生成文件和上下传代码
	genCode := router.Group("dfcCode")
	{
		genCode.POST("generateCode", controller.HandleGenerateCode)
		genCode.GET("/", controller.GetTest)
		genCode.POST("uploadFile", controller.HandleSendFileWithSSH)
		genCode.POST("downloadFile", controller.HandleDownloadFileWithSSH)
	}
	// 以下两个接口用来做文件服务器
	// 两个接口：保存文件和下载文件
	fileOpera := router.Group("fileOpera")
	{
		fileOpera.GET("downloadFile", controller.HandleDownloadFile)
		fileOpera.POST("storeFile", controller.HandleSaveFile)
	}
}

func initDirs() {
	dirNeedToInit := []string{"./log", "./codeList", "./dfcCodeTransit"}
	for _, item := range dirNeedToInit {
		succ, err := initDir(item)
		if !succ {
			log.Fatal("make dir err", err.Error())
		}
	}
}

func initDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		errMkDir := os.Mkdir(path, os.ModePerm)
		return errMkDir == nil, errMkDir
	}
	return true, nil
}

func Run() {
	initDirs()
	r := initRouter()
	r.Run(":8000")
}
