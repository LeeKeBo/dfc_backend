package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func PrintInLog(format string, a ...interface{}){
	fmt.Fprintf(gin.DefaultWriter,format,a)
}
