package controller

import (
	"dfc_backend/app/model"
	"dfc_backend/app/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// 设置一个提供字符的默认数组，用来生成随机字符串
var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var filePosition = "./codeList/"

func HandleGenerateCode(ctx *gin.Context) {
	var codeData model.CodeData
	err := ctx.Bind(&codeData)
	if err != nil {
		util.PrintInLog("bind data:%+v,error:%s\n", codeData, err.Error())

	} else {
		// 这里来添加文件
		fileName := filePosition + GetRandomStr(8) + ".c"
		codeFile, err := os.Create(fileName)
		defer codeFile.Close()
		if err != nil {
			util.PrintInLog("get codeFile error:%+s\n", err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "生成失败",
			})
			return
		} else {
			code, err := GetCodeContent(&codeData)
			if err != nil {
				util.PrintInLog("get code content error:%+s\n", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "生成失败",
				})
				return
			}
			_, err = codeFile.WriteString(code)
			if err != nil {
				util.PrintInLog("write file error:%+s\n", err.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "生成失败",
				})
			}
			ctx.File(fileName)
			//fileContentDisposition := "attachment;filename=\"" + "dfcCodeFile.c" + "\""
			//ctx.Header("Content-Type", "application/text/plain") // 这里是压缩文件类型 .zip
			//ctx.Header("Content-Disposition", fileContentDisposition)
			//ctx.Header("Accept-Length", fmt.Sprintf("%d", len(code)))
			//ctx.Writer.Write([]byte(code))
		}
		util.PrintInLog("codeData:%+v\n", codeData)
	}
}

func GetCodeContent(codeData *model.CodeData) (string, error) {
	if codeData == nil {
		return "", fmt.Errorf("nil node")
	}
	var (
		nodeString     string
		tempNodeString string
		graphString    string
		err            error
	)
	// 这里是ID到节点名称对应的映射，这里的ID包括节点的ID，也包括typeId
	mapNodeIdToName := make(map[string]string)
	for _, node := range codeData.ChartData.Nodes {
		mapNodeIdToName[node.ID] = node.Name
	}
	for _, subGraphNode := range codeData.NodeList[model.NODE_TYPE_SUBGRAPHNODE] {
		mapNodeIdToName[subGraphNode.ID] = subGraphNode.Name
		for _, node := range subGraphNode.Nodes {
			mapNodeIdToName[node.ID] = node.Name
		}
	}
	for _, funcNode := range codeData.NodeList[model.NODE_TYPE_FLOWNODE] {
		mapNodeIdToName[funcNode.ID] = funcNode.Name
	}

	graphString, err = GetCodeFromGraphNode(&codeData.ChartData, mapNodeIdToName, 1)
	if err != nil {
		util.PrintInLog("get Code error:%+s\n", err.Error())
		//fmt.Println()
		return "", err
	}
	for _, nodeList := range codeData.NodeList {
		for _, node := range nodeList {
			tempNodeString, err = GetCodeFromNode(&node, mapNodeIdToName, 0)
			if err != nil {
				util.PrintInLog("get code from node error:%s\n", err.Error())
			}
			if node.Type == model.NODE_TYPE_SUBGRAPHNODE {
				graphString += tempNodeString
			} else {
				nodeString += tempNodeString
			}
		}
	}
	graphString = "void graph(){\n" + graphString + "}\n"
	return nodeString + graphString, nil
	//codeContent.WriteString("")
	//return codeContent.String(), nil
}

func GetCodeFromNode(node *model.Node, mapNodeIdToName map[string]string, spaceNum int) (string, error) {
	var codeContent strings.Builder
	nodeType := node.Type
	if nodeType == model.NODE_TYPE_STRUCTNODE {
		str := "struct " + node.Name + "{\n"
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + str)
		// Todo:添加结构体变量,这里还有结构体可见性未解决
		spaceNum++
		for _, attr := range node.Attrs {
			codeContent.WriteString(strings.Repeat("\t", spaceNum) + attr.Type + " " + attr.Name + ";\n")
		}
		// 添加结构体函数
		funcCode := strings.Split(node.Code, "\n")
		for _, codeLine := range funcCode {
			codeContent.WriteString(strings.Repeat("\t", spaceNum) + codeLine + "\n")
		}
		spaceNum--
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + "};\n")
	} else if nodeType == model.NODE_TYPE_FLOWNODE {
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + node.ReturnType + " " + node.Name + "(")
		for index, attr := range node.Inputs {
			codeContent.WriteString(attr.Type + " " + attr.Name)
			if index != len(node.Inputs)-1 {
				codeContent.WriteString(",")
			}
		}
		// 如果两种参数都没有，那么就是没有参数，不需加分号分隔
		if len(node.Inputs) != 0 || len(node.Outputs) != 0 {
			codeContent.WriteString(";")
		}
		for index, attr := range node.Outputs {
			codeContent.WriteString(attr.Type + " " + attr.Name)
			if index != len(node.Outputs)-1 {
				codeContent.WriteString(",")
			}
		}
		codeContent.WriteString("){\n")
		spaceNum++
		funcCode := strings.Split(node.Code, "\n")
		for _, codeLine := range funcCode {
			codeContent.WriteString(strings.Repeat("\t", spaceNum) + codeLine + "\n")
		}
		spaceNum--
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + "}\n")
	} else if nodeType == model.NODE_TYPE_SUBGRAPHNODE {
		return GetCodeFromGraphNode(node, mapNodeIdToName, spaceNum+1)
	}
	return codeContent.String(), nil
}

func GetCodeFromGraphNode(node *model.Node, mapNodeIdToName map[string]string, spaceNum int) (string, error) {
	var codeContent strings.Builder
	codeContent.WriteString(strings.Repeat("\t", spaceNum))
	if node.Type == model.NODE_TYPE_SUBGRAPHNODE {
		codeContent.WriteString("subgraph " + node.Name + "(")
	} else {
		codeContent.WriteString("graph " + node.Name)
	}
	for index, attr := range node.Inputs {
		codeContent.WriteString(attr.Type + " " + attr.Name)
		if index != len(node.Inputs)-1 {
			codeContent.WriteString(",")
		}
	}
	// 只有子图节点需要记录这个信息，主图没有输入输出，所有没有分号来分隔输入输出
	if node.Type == model.NODE_TYPE_SUBGRAPHNODE {
		codeContent.WriteString(";")
	}
	for index, attr := range node.Outputs {
		codeContent.WriteString(attr.Type + " " + attr.Name)
		if index != len(node.Inputs)-1 {
			codeContent.WriteString(",")
		}
	}
	if node.Type == model.NODE_TYPE_SUBGRAPHNODE {
		codeContent.WriteString(")")
	}
	codeContent.WriteString("{\n")
	spaceNum++
	for _, FNNode := range node.Nodes {
		if FNNode.Type == model.NODE_TYPE_FLOWNODE {
			if _, ok := mapNodeIdToName[FNNode.ID]; ok {
				codeContent.WriteString(strings.Repeat("\t", spaceNum) + "FN " + FNNode.Name + " " + mapNodeIdToName[FNNode.TypeId] + ";\n")
			}
		} else if FNNode.Type == model.NODE_TYPE_SUBGRAPHNODE {
			if _, ok := mapNodeIdToName[FNNode.ID]; ok {
				codeContent.WriteString(strings.Repeat("\t", spaceNum) + "subgraph " + FNNode.Name + " " + mapNodeIdToName[FNNode.TypeId] + ";\n")
			}
		}
	}
	for _, ADNode := range node.Connections {
		if len(ADNode.Attrs) == 0 {
			continue
		}
		// 前端传过来的input和output格式如下: int xxx (0) （括号内表示是第几个参数）
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + "AD " + strconv.Itoa(len(ADNode.Attrs)) + " " + mapNodeIdToName[ADNode.SourceId])
		for _, attr := range ADNode.Attrs {
			input := attr.Input[strings.LastIndex(attr.Input, "(")+1 : strings.LastIndex(attr.Input, ")")]
			codeContent.WriteString(" " + input)
		}
		codeContent.WriteString(" " + mapNodeIdToName[ADNode.TargetId])
		for _, attr := range ADNode.Attrs {
			output := attr.Output[strings.LastIndex(attr.Output, "(")+1 : strings.LastIndex(attr.Output, ")")]
			codeContent.WriteString(" " + output)
		}
		codeContent.WriteString(";\n")
	}
	spaceNum--
	codeContent.WriteString(strings.Repeat("\t", spaceNum) + "}\n")
	return codeContent.String(), nil
}

func GetRandomStr(strLength int) string {
	randomStr := make([]rune, strLength)
	for i := range randomStr {
		randomStr[i] = defaultLetters[rand.Intn(len(defaultLetters))]
	}
	return string(randomStr)
}

func GetTest(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status":  "200",
		"message": "生成成功",
	})
}
