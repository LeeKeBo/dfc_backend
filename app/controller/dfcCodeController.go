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
	"sync"
)

// 设置一个提供字符的默认数组，用来生成随机字符串
var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var CodeFilePosition = "./codeList/"

func HandleGenerateCode(ctx *gin.Context) {
	var codeData model.CodeData
	err := ctx.Bind(&codeData)
	if err != nil {
		util.PrintInLog("bind data:%+v,error:%s\n", codeData, err.Error())

	} else {
		codeList, err := GetCodeContent(&codeData)
		if err != nil || codeList == nil || len(codeList) != 2 {
			util.PrintInLog("get code content error:%+s\n", err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "生成失败",
			})
			return
		}
		errChan := make(chan error)
		defer close(errChan)
		wg := sync.WaitGroup{}
		wg.Add(2)
		// 这里来添加文件
		randomStr := GetRandomStr(8)
		codeFileName := CodeFilePosition + randomStr + ".c"
		graphFileName := CodeFilePosition + randomStr + ".graph"
		go func() {
			codeFile, err := os.Create(codeFileName)
			defer codeFile.Close()
			_, err = codeFile.WriteString(codeList[0])
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
		go func() {
			graphFile, err := os.Create(graphFileName)
			defer graphFile.Close()
			_, err = graphFile.WriteString(codeList[1])
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
		wg.Wait()
		select {
		case err = <-errChan:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"err": err.Error(),
			})
		default:
			ctx.JSON(http.StatusOK, gin.H{
				"graphCode": randomStr + ".graph",
				"dfcCode":   randomStr + ".c",
			})
		}
		util.PrintInLog("codeData:%+v\n", codeData)
	}
}

func GetCodeContent(codeData *model.CodeData) ([]string, error) {
	if codeData == nil {
		return nil, fmt.Errorf("nil node")
	}
	var (
		nodeString     string
		tempNodeString string
		graphString    string
		err            error
	)
	// 这里是ID到节点名称对应的映射，这里的ID包括节点的ID，也包括typeId
	mapNodeIdToNodeTemp := make(map[string]model.Node)
	for _, node := range codeData.ChartData.Nodes {
		mapNodeIdToNodeTemp[node.ID] = node
	}
	for _, subGraphNode := range codeData.NodeList[model.NODE_TYPE_SUBGRAPHNODE] {
		mapNodeIdToNodeTemp[subGraphNode.ID] = subGraphNode
		for _, node := range subGraphNode.Nodes {
			mapNodeIdToNodeTemp[node.ID] = node
		}
	}
	for _, funcNode := range codeData.NodeList[model.NODE_TYPE_FLOWNODE] {
		mapNodeIdToNodeTemp[funcNode.ID] = funcNode
	}

	graphString, err = GetCodeFromGraphNode(&codeData.ChartData, &mapNodeIdToNodeTemp, 0)
	if err != nil {
		util.PrintInLog("get Code error:%+s\n", err.Error())
		//fmt.Println()
		return nil, err
	}
	for _, nodeList := range codeData.NodeList {
		for _, node := range nodeList {
			tempNodeString, err = GetCodeFromNode(&node, &mapNodeIdToNodeTemp, 0)
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
	//graphString = "void graph(){\n" + graphString + "}\n"
	return []string{nodeString, graphString}, nil
	//codeContent.WriteString("")
	//return codeContent.String(), nil
}

func GetCodeFromNode(node *model.Node, mapNodeIdToNodeTemp *map[string]model.Node, spaceNum int) (string, error) {
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
		return GetCodeFromGraphNode(node, mapNodeIdToNodeTemp, spaceNum)
	}
	return codeContent.String(), nil
}

func GetCodeFromGraphNode(node *model.Node, mapNodeIdToNodeTemp *map[string]model.Node, spaceNum int) (string, error) {
	var codeContent strings.Builder
	codeContent.WriteString(strings.Repeat("\t", spaceNum))
	codeContent.WriteString("GRAPH " + node.Name + "\n{\n")
	// 不需要记录子图参数信息，就先注释吧，需要了再取消即可
	//for index, attr := range node.Inputs {
	//	codeContent.WriteString(attr.Type + " " + attr.Name)
	//	if index != len(node.Inputs)-1 {
	//		codeContent.WriteString(",")
	//	}
	//}
	//// 只有子图节点需要记录这个信息，主图没有输入输出，所有没有分号来分隔输入输出
	//if node.Type == model.NODE_TYPE_SUBGRAPHNODE {
	//	codeContent.WriteString(";")
	//}
	//for index, attr := range node.Outputs {
	//	codeContent.WriteString(attr.Type + " " + attr.Name)
	//	if index != len(node.Inputs)-1 {
	//		codeContent.WriteString(",")
	//	}
	//}
	spaceNum++
	for _, FNNode := range node.Nodes {
		nodeTemp := (*mapNodeIdToNodeTemp)[FNNode.TypeId]
		if FNNode.Type == model.NODE_TYPE_FLOWNODE {
			if _, ok := (*mapNodeIdToNodeTemp)[FNNode.ID]; ok {
				codeContent.WriteString(strings.Repeat("\t", spaceNum) + "FN " + FNNode.Name + " " + nodeTemp.Name + " N;\n")
			}
		} else if FNNode.Type == model.NODE_TYPE_SUBGRAPHNODE {
			if _, ok := (*mapNodeIdToNodeTemp)[FNNode.ID]; ok {
				codeContent.WriteString(strings.Repeat("\t", spaceNum) + "GRAPH " + FNNode.Name + " " + nodeTemp.Name +
					" " + strconv.Itoa(len(nodeTemp.Inputs)) + " " + strconv.Itoa(len(nodeTemp.Outputs)) + " N;\n")
				fmt.Println(FNNode.Inputs)
				fmt.Println(FNNode.Outputs)
			}
		}
	}
	for _, ADNode := range node.Connections {
		if len(ADNode.Attrs) == 0 {
			continue
		}
		// 前端传过来的input和output格式如下: int xxx (0) （括号内表示是第几个参数）
		codeContent.WriteString(strings.Repeat("\t", spaceNum) + "AD " + strconv.Itoa(len(ADNode.Attrs)) + " " + (*mapNodeIdToNodeTemp)[ADNode.SourceId].Name)
		for _, attr := range ADNode.Attrs {
			input := attr.Input[strings.LastIndex(attr.Input, "(")+1 : strings.LastIndex(attr.Input, ")")]
			codeContent.WriteString(" " + input)
		}
		codeContent.WriteString(" " + (*mapNodeIdToNodeTemp)[ADNode.TargetId].Name)
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
