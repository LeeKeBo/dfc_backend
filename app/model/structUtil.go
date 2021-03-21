package model

const (
	NODE_TYPE_STRUCTNODE = iota
	NODE_TYPE_FLOWNODE
	NODE_TYPE_SUBGRAPHNODE
)

// Attr 是用来表示一个变量的属性
type Attr struct {
	Visibility string `json:"visibility"` // eg: public,protect,private
	Type       string `json:"type"`
	Name       string `json:"name"`
}

type ConnAttr struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Connection 表示一条连线
type Connection struct {
	ID           string     `json:"id"`
	SourceId     string     `json:"sourceId"`
	SourceTypeId string     `json:"sourceTypeId"`
	SourceType   int        `json:"sourceType"`
	TargetId     string     `json:"targetId"`
	TargetTypeId string     `json:"targetTypeId"`
	TargetType   int        `json:"targetType"`
	Text         string     `json:"text"`
	Attrs        []ConnAttr `json:"attrs"`
}

// Node 表示一个节点，这里整个图也可以看成一个节点，所以节点也会有记录内部的连线情况的连线数组
type Node struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        int          `json:"type"`
	TypeId      string       `json:"typeId"`
	Inputs      []Attr       `json:"inputs"`
	Outputs     []Attr       `json:"outputs"`
	Attrs       []Attr       `json:"attrs"`
	Code        string       `json:"code"`
	Nodes       []Node       `json:"nodes"`
	Connections []Connection `json:"connections"`
	ReturnType  string       `json:"returnType"`
}

// 主图信息
type ChartData struct {
	Nodes       []Node       `json:"nodes"`
	Connections []Connection `json:"connections"`
}

// 接收的生成代码相关的结构体
type CodeData struct {
	ChartData Node     `json:"chartData"`
	NodeList  [][]Node `json:"nodeList"`
}
