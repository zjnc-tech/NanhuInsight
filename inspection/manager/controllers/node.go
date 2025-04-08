package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/xuri/excelize/v2"

	"infrahi/backend/inspection-manager/api"
	"infrahi/backend/inspection-manager/pkg/cluster"
)

type NodeController struct {
	beego.Controller
}

type UploadNodes struct {
	Ready            []string `json:"ready"`
	NotReady         []string `json:"notReady"`
	Unschedulable    []string `json:"unschedulable"`
	Allocated        []string `json:"allocated"`
	Master           []string `json:"master"`
	ResourceNotMatch []string `json:"resourceNotMatch"`
	IPNotMatch       []string `json:"nodeNotMatch"`
}

type UploadResult struct {
	UploadNodes UploadNodes `json:"uploadNodes"`
	TotalNum    int         `json:"totalNum"`
	AnalyzedNum int         `json:"analyzedNum"`
	SuccessNum  int         `json:"successNum"`
}

// GetAll ...
// @Summary     获取集群节点
// @Description 根据集群名称,卡资源类型和模式，获取key为节点类型，value为节点IP列表的map
// @Tags        节点管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于指定要查询节点的集群"
// @Param       mode      query   string true "过滤模式，用于指定例行还是深度检查"
// @Param       resource  query   string true "资源名称，用于指定要查询的节点"
// @Success     200 {object} api.CommonResponse "成功返回包含节点 IP 地址列表的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数（如集群名称、模式或资源）"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/nodes [get]
func (c *NodeController) GetAll() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	mode := c.GetString("mode")
	resource := c.GetString("resource")

	if clusterName == "" || mode == "" || resource == "" {
		c.Data["json"] = api.ParamErrResponse("clusterName, mode, and resource are required")
		c.ServeJSON()
		return
	}

	IPMap, err := cluster.GetNodeStatusMap(clusterName, mode, resource)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	delete(IPMap, "master")

	resourceMap, err := cluster.GetResourceMap(clusterName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	resourceList, exists := resourceMap[resource]
	if !exists {
		c.Data["json"] = api.InternalErrResponse(fmt.Sprintf("Nodes of resource %s not found", resource))
		c.ServeJSON()
		return
	}

	IPListSet := make(map[string]struct{})
	for _, ip := range resourceList {
		IPListSet[ip] = struct{}{}
	}

	// 遍历 ipMap，并对每个分类进行过滤
	for category, ips := range IPMap {
		var filteredIPs []string
		for _, ip := range ips {
			if _, exists := IPListSet[ip]; exists {
				filteredIPs = append(filteredIPs, ip)
			}
		}

		// 更新分类中的 IP 列表
		IPMap[category] = filteredIPs
	}

	c.Data["json"] = api.SuccessResponse(IPMap)
	c.ServeJSON()
}

// GetAllOld ...
// @Summary     获取集群节点
// @Description 根据集群名称和模式，获取节点（IP 地址）列表，并根据资源名返回相关节点信息
// @Tags        节点管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于指定要查询节点的集群"
// @Param       mode      query   string true "过滤模式，用于根据特定条件筛选节点"
// @Param       resource  query   string true "资源名称，用于指定要查询的节点"
// @Success     200 {object} api.CommonResponse "成功返回包含节点 IP 地址列表的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数（如集群名称、模式或资源）"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/get_nodes_from_cluster [get]
func (c *NodeController) GetAllOld() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	mode := c.GetString("mode")
	resource := c.GetString("resource")

	if clusterName == "" || mode == "" || resource == "" {
		c.Data["json"] = api.ParamErrResponse("clusterName, mode, and resource are required")
		c.ServeJSON()
		return
	}

	resourceMap, err := cluster.GetResourceMap(clusterName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	if data, exists := resourceMap[resource]; exists {
		c.Data["json"] = api.SuccessResponse(data)
	} else {
		c.Data["json"] = api.InternalErrResponse(fmt.Sprintf("Nodes of resource %s not found", resource))
	}

	c.ServeJSON()
}

// Resource ...
// @Summary     获取集群资源
// @Description 根据集群名称获取集群资源列表
// @Tags        节点管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于指定要查询节点的集群"
// @Success     200 {object} api.CommonResponse "成功返回包含节点 IP 地址列表的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数（如集群名称或模式）或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/node/resource [get]
func (c *NodeController) Resource() {
	clusterName := c.Ctx.Input.Header("x-cluster")

	if clusterName == "" {
		c.Data["json"] = api.ParamErrResponse("clusterName is required")
		c.ServeJSON()
		return
	}

	resources, err := beego.AppConfig.String(clusterName + "::resource")
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	resourceList := strings.Split(resources, ",")
	//resourceMap, err := cluster.GetResourceMap(clusterName, "")
	//if err != nil {
	//	c.Data["json"] = api.InternalErrResponse(err.Error())
	//	c.ServeJSON()
	//	return
	//}
	//
	//resources := make([]string, 0, len(resourceMap))
	//for key := range resourceMap {
	//	// 未识别不返给前端
	//	if key == "Unidentified" {
	//		continue
	//	}
	//	resources = append(resources, key)
	//}

	c.Data["json"] = api.SuccessResponse(resourceList)
	c.ServeJSON()
}

// DownloadExcel ...
// @Summary     下载 Excel 模板文件
// @Description 下载节点表格模板
// @Tags        节点管理
// @Accept      json
// @Produce     application/octet-stream
// @Success     200 {string} file "成功节点 Excel 文件"
// @Failure     400 {object} api.CommonResponse "请求错误，例如文件不存在"
// @Failure     500 {object} api.CommonResponse "服务器内部错误，例如文件路径读取失败"
// @Router      /inspection/api/v1/node/download_excel [get]
func (c *NodeController) DownloadExcel() {
	// 构建文件路径
	tempPath, _ := beego.AppConfig.String(beego.BConfig.RunMode + "::kube_path")
	filePath := filepath.Join(tempPath, "nodes.xlsx")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.Data["json"] = api.ParamErrResponse("nodes.xlsx not exist")
		c.ServeJSON()
		return
	}

	// 发送文件
	c.Ctx.Output.Download(filePath)
}

// UploadExcel ...
// @Summary     上传 Excel 文件
// @Description 上传包含 IP 地址列表的 Excel 文件，与指定集群的节点进行比较，并返回匹配和未匹配结果。
// @Tags        节点管理
// @Accept      multipart/form-data
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于指定要比较的集群"
// @Param       file      formData file   true "上传的 Excel 文件（支持 .xlsx 和 .xls 格式）"
// @Param       mode      query    string true "过滤模式，用于指定例行还是深度检查"
// @Param       resource  query    string true "资源名称，用于指定要比较的资源"
// @Success     200 {object} api.CommonResponse "成功返回匹配的 IP 列表和匹配统计信息"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少集群名称、文件参数或资源名称"
// @Failure     500 {object} api.CommonResponse "服务器内部错误，例如文件解析失败或处理逻辑错误"
// @Router      /inspection/api/v1/node/upload_excel [post]
func (c *NodeController) UploadExcel() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	mode := c.GetString("mode")
	resource := c.GetString("resource")

	if clusterName == "" || resource == "" || mode == "" {
		c.Data["json"] = api.ParamErrResponse("Param clusterName, mode and resource are required")
		c.ServeJSON()
		return
	}

	// 获取上传的文件
	file, header, err := c.GetFile("file")
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}
	defer file.Close()

	// 确保上传的是 Excel 文件
	if !strings.HasSuffix(header.Filename, ".xlsx") && !strings.HasSuffix(header.Filename, ".xls") {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	tempPath, _ := beego.AppConfig.String(beego.BConfig.RunMode + "::script_path")

	// 保存文件到临时目录
	uploadPath := filepath.Join(tempPath, header.Filename)

	if err = c.SaveToFile("file", uploadPath); err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}
	defer os.Remove(uploadPath) // 删除临时文件

	// 读取 Excel 文件内容
	IPList, err := readExcel(uploadPath)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	IPMap, err := cluster.GetNodeStatusMap(clusterName, mode, resource)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	reverseIPMap := make(map[string]string)
	for category, ips := range IPMap {
		for _, ip := range ips {
			reverseIPMap[ip] = category
		}
	}

	resourceMap, err := cluster.GetResourceMap(clusterName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	nodes := uploadNodes(resourceMap, reverseIPMap, resource, IPList)

	analyzedNum := len(IPList) - len(nodes.IPNotMatch)
	if analyzedNum < 0 {
		analyzedNum = 0
	}

	result := UploadResult{
		UploadNodes: nodes,
		TotalNum:    len(IPList),
		AnalyzedNum: analyzedNum,
		SuccessNum:  len(nodes.Ready),
	}

	c.Data["json"] = api.SuccessResponse(result)
	c.ServeJSON()
}

// readExcel 解析 Excel 文件并返回文件内的 IP 地址列表
func readExcel(filePath string) ([]string, error) {
	fmt.Printf("filePath:%s\n", filePath)
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	// 获取第一个 Sheet 名
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	// 读取第一列数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	// 提取 IP 列表
	var ipList []string
	for i, row := range rows {
		if i == 0 { // 跳过第一行（表头）
			continue
		}
		if len(row) > 0 {
			ip := strings.TrimSpace(row[0]) // 获取第一列的值
			if ip != "" {
				ipList = append(ipList, ip)
			}
		}
	}

	return ipList, nil
}

func uploadNodes(resourceMap map[string][]string, reverseIPMap map[string]string,
	resource string, excelList []string) UploadNodes {
	nodes := UploadNodes{}

	resourceIPs := resourceMap[resource]
	allMapSet := make(map[string]bool)
	for _, ips := range resourceMap {
		for _, ip := range ips {
			allMapSet[ip] = true
		}
	}

	for _, ip := range excelList {
		if contains(resourceIPs, ip) {
			if reverseIPMap[ip] == "ready" {
				nodes.Ready = append(nodes.Ready, ip)
			} else if reverseIPMap[ip] == "unschedulable" {
				nodes.Unschedulable = append(nodes.Unschedulable, ip)
			} else if reverseIPMap[ip] == "allocated" {
				nodes.Allocated = append(nodes.Allocated, ip)
			} else if reverseIPMap[ip] == "notReady" {
				nodes.NotReady = append(nodes.NotReady, ip)
			}
		} else if reverseIPMap[ip] == "master" {
			nodes.Master = append(nodes.Master, ip)
		} else if allMapSet[ip] {
			nodes.ResourceNotMatch = append(nodes.ResourceNotMatch, ip)
		} else {
			nodes.IPNotMatch = append(nodes.IPNotMatch, ip)
		}
	}
	return nodes
}

// 辅助函数：检查切片中是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func intersection(list1, list2 []string) ([]string, int, int) {
	matches := 0
	nonMatches := 0

	// 创建一个 map 来存储第一个列表中的 IP 地址
	ipMap := make(map[string]struct{})
	var result []string

	// 将第一个列表的 IP 地址存储到 map 中
	for _, ip := range list1 {
		ipMap[ip] = struct{}{}
	}

	// 遍历第二个列表，检查每个 IP 是否在 map 中
	for _, ip := range list2 {
		if _, exists := ipMap[ip]; exists {
			result = append(result, ip)
			matches++
		} else {
			nonMatches++
		}
	}

	return result, matches, nonMatches
}
