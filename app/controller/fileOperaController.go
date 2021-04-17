package controller

import (
	"dfc_backend/app/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path"
	"time"
)

// 上传文件功能是为了文件SSH服务
func HandleSaveFile(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.String(http.StatusBadRequest, "接收文件失败")
	}
	util.PrintInLog("file route:%s\n", FilePre+file.Filename)
	if err := ctx.SaveUploadedFile(file, FilePre+file.Filename); err != nil {
		util.PrintInLog("save file error:%s\n", err.Error())
		ctx.String(http.StatusBadRequest, "保存文件失败")
	}
}

// 文件下载为为了代码下载服务
func HandleDownloadFile(ctx *gin.Context) {
	fmt.Println(ctx.Request.URL.Query())
	fileName := ctx.Query("fileName")
	//打开文件,这里会限定只能下载CodeFilePosition下的文件
	_, errByOpenFile := os.Open(CodeFilePosition + path.Base(fileName))
	fmt.Println("get filePath:", CodeFilePosition+path.Base(fileName))
	//非空处理
	if errByOpenFile != nil {
		/*c.JSON(http.StatusOK, gin.H{
		    "success": false,
		    "message": "失败",
		    "error":   "资源不存在",
		})*/
		ctx.JSON(http.StatusFound, gin.H{})
		return
	}
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+time.Now().Format("2006.01.02 15:04:05")+path.Ext(fileName))
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.File(CodeFilePosition + path.Base(fileName))
}
