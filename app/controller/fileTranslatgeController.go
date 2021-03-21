// 这里是先创建SSH连接再进行文件传输，主要使用了SSH库和github.com/pkg/sftp这个包
package controller

import (
	"dfc_backend/app/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

type SSHInfo struct {
	user     string
	password string
	host     string
	port     int
}

type FormData struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	RemoteDir  string `json:"remoteDir"`
	File       string `json:"file"`
	RemoteFile string `json:"remoteFile"`
}

var FilePre string = "./dfcCodeTransit/"

// 这是创建SSH连接的函数
func createConnection(sshInfo SSHInfo) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(sshInfo.password))

	clientConfig = &ssh.ClientConfig{
		User:    sshInfo.user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr = fmt.Sprintf("%s:%d", sshInfo.host, sshInfo.port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	return sftpClient, nil
}

func sendFile(sshInfo SSHInfo, localFilePath, remoteDir string) {
	var (
		err        error
		sftpClient *sftp.Client
	)
	sftpClient, err = createConnection(sshInfo)
	if err != nil {
		util.PrintInLog("create conn error:%s\n", err.Error())
		return
	}
	defer sftpClient.Close()

	// 源文件
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		util.PrintInLog("open file error:%s\n", err.Error())
		return
	}
	defer srcFile.Close()

	// 目标文件
	remoteFileName := path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		util.PrintInLog("create file error:%s\n", err.Error())
		return
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		byteNum, _ := srcFile.Read(buf)
		if byteNum == 0 {
			break
		}
		dstFile.Write(buf)
	}
	//util.PrintInLog("succ\n")
}

func getFile(sshInfo SSHInfo, remoteFilePath, localDir string) {
	var (
		err        error
		sftpClient *sftp.Client
	)
	sftpClient, err = createConnection(sshInfo)
	if err != nil {
		util.PrintInLog("create conn error:%s\n", err.Error())
		return
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		util.PrintInLog("open file error:%s\n", err.Error())
		return
	}
	defer srcFile.Close()

	localFileName := path.Base(remoteFilePath)
	dstFile, err := os.Create(path.Join(localDir, localFileName))
	if err != nil {
		util.PrintInLog("create file error:%s\n", err.Error())
		return
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		util.PrintInLog("write file error:%s\n", err.Error())
		return
	}
	//fmt.Println("copy file finished!")
}

func HandleSendFile(ctx *gin.Context) {
	//fmt.Println(ctx.Po)
	var formData FormData
	err := ctx.Bind(&formData)
	if err != nil {
		util.PrintInLog("bind data: %+v, error:%s\n", formData, err.Error())
		ctx.String(http.StatusBadRequest, "Bad request")
		return
	}
	//fmt.Println(formData)
	var sshInfo = SSHInfo{
		user:     formData.Name,
		password: formData.Password,
		host:     formData.Host,
		port:     formData.Port,
	}
	remoteDir := formData.RemoteDir
	localFilePath := FilePre + formData.File
	sendFile(sshInfo, localFilePath, remoteDir)
	ctx.String(http.StatusOK, "ok")

}

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

func HandleDownloadFile(ctx *gin.Context) {
	var formData FormData
	err := ctx.Bind(&formData)
	if err != nil {
		util.PrintInLog("bind data: %+v, error:%s\n", formData, err.Error())
		ctx.String(http.StatusBadRequest, "Bad request")
		return
	}
	//fmt.Println(formData)
	var sshInfo = SSHInfo{
		user:     formData.Name,
		password: formData.Password,
		host:     formData.Host,
		port:     formData.Port,
	}
	remoteFile := formData.RemoteFile

	//localFilePath := FilePre + formData.File
	getFile(sshInfo, remoteFile, FilePre)
	ctx.Header("fileName", path.Base(remoteFile))
	ctx.File(FilePre + path.Base(remoteFile))
}
