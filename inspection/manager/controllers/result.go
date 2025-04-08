package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/api"
	"infrahi/backend/inspection-manager/models"
)

type TestInfoController struct {
	beego.Controller
}

// DownloadRequest represents the request body containing jobId and caseName
type DownloadRequest struct {
	JobId      string `json:"jobId"`
	CaseChName string `json:"caseName"`
}

// Get ...
// @Summary     获取测试结果详情
// @Description 根据作业 ID 获取测试结果的详细信息
// @Tags        检查结果
// @Accept      json
// @Produce     json
// @Param       jobId query   string true "用于查询检查结果的作业ID"
// @Success     200 {object} api.CommonResponse "成功返回检查结果详情"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少 jobId 参数"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/result/summary [get]
func (c *TestInfoController) Get() {
	jobId := c.GetString("jobId")

	if jobId == "" {
		c.Data["json"] = api.ParamErrResponse("jobId is required")
		c.ServeJSON()
		return
	}

	detail, err := getTestResultDetail(jobId)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(detail)
	c.ServeJSON()
}

// GetOld ...
// @Summary     获取测试结果详情
// @Description 根据作业 ID 获取测试结果的详细信息
// @Tags        检查结果
// @Accept      json
// @Produce     json
// @Param       jobId query   string true "用于查询检查结果的作业ID"
// @Success     200 {object} api.CommonResponse "成功返回检查结果详情"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少 jobId 参数"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/test_result_details [get]
func (c *TestInfoController) GetOld() {
	jobId := c.GetString("jobId")

	if jobId == "" {
		c.Data["json"] = api.ParamErrResponse("jobId is required")
		c.ServeJSON()
		return
	}

	detail, err := getTestResultDetail(jobId)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(detail)
	c.ServeJSON()
}

// GetOne ...
// @Summary     获取特定测试用例详情
// @Description 根据作业 ID 和用例名称获取某个测试用例的详细信息
// @Tags        检查结果
// @Accept      json
// @Produce     json
// @Param       jobId    query   string true "用于标识测试的作业 ID"
// @Param       caseName query   string true "指定要查询的测试用例名称"
// @Success     200 {object} api.CommonResponse "成功返回特定测试用例的详情"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少 jobId 或 caseName 参数"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/result/detail [get]
func (c *TestInfoController) GetOne() {
	jobId := c.GetString("jobId")
	caseChName := c.GetString("caseName")

	if jobId == "" || caseChName == "" {
		c.Data["json"] = api.ParamErrResponse("jobId and caseName are required")
		c.ServeJSON()
		return
	}

	detail, err := getScriptDetailInfo(jobId, caseChName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(detail)
	c.ServeJSON()
}

// GetOneOld ...
// @Summary     获取特定测试用例详情
// @Description 根据作业 ID 和用例名称获取某个测试用例的详细信息
// @Tags        检查结果
// @Accept      json
// @Produce     json
// @Param       jobId    query   string true "用于标识测试的作业 ID"
// @Param       caseName query   string true "指定要查询的测试用例名称"
// @Success     200 {object} api.CommonResponse "成功返回特定测试用例的详情"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少 jobId 或 caseName 参数"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/test_case_detail [get]
func (c *TestInfoController) GetOneOld() {
	jobId := c.GetString("jobId")
	caseChName := c.GetString("caseName")

	if jobId == "" || caseChName == "" {
		c.Data["json"] = api.ParamErrResponse("jobId and caseName are required")
		c.ServeJSON()
		return
	}

	detail, err := getScriptDetailInfo(jobId, caseChName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(detail)
	c.ServeJSON()
}

// Post ...
// @Summary     下载测试用例日志
// @Description 下载指定作业中的检查结果的日志文件
// @Tags        检查结果
// @Accept      json
// @Produce     json
// @Param       body body DownloadRequest true "包含 jobId 和 caseName 的 JSON 对象"
// @Success     200 {file} string "指定测试用例的日志文件"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少或无效的参数"
// @Failure     404 {object} api.CommonResponse "未找到日志文件"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/download_log [post]
func (c *TestInfoController) Post() {
	var request DownloadRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.Data["json"] = api.ParamErrResponse("Invalid JSON payload")
		c.ServeJSON()
		return
	}

	if request.JobId == "" || request.CaseChName == "" {
		c.Data["json"] = api.ParamErrResponse("jobId and caseName are required")
		c.ServeJSON()
		return
	}

	// 获取 case_name
	scriptName, err := models.QueryScriptNameByChName(request.CaseChName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	// 构建文件路径
	logPath, _ := beego.AppConfig.String("log_path")
	filePath := filepath.Join(logPath, fmt.Sprintf("%s", request.JobId), scriptName+".log")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.Data["json"] = api.ParamErrResponse("no log exist")
		c.ServeJSON()
		return
	}

	// 发送文件
	c.Ctx.Output.Download(filePath)
}

func getTestResultDetail(jobId string) (interface{}, error) {
	jobInfo, err := models.GetJobInfo(jobId)
	if err != nil {
		return nil, err
	}

	records, err := models.GetTestRecords(jobId)
	if err != nil {
		return nil, err
	}
	createTime, _ := models.QueryJobInfoById(jobId, "create_time")
	passedNum := models.QueryCountByTestResult(jobId, "passed")
	notPassedNum := models.QueryCountByTestResult(jobId, "not passed")
	failedNum := models.QueryCountByTestResult(jobId, "failed")

	result := map[string]interface{}{
		"jobName":          jobInfo.JobName,
		"templateName":     jobInfo.TemplateName,
		"jobStatus":        jobInfo.Status,
		"resource":         jobInfo.Resource,
		"IPList":           models.ParseIPList(jobInfo.IpList),
		"baseIP":           jobInfo.BaseIp,
		"timeCost":         getJobTimeCost(jobId),
		"createTime":       createTime,
		"totalCasesNum":    len(records),
		"notPassedCaseNum": notPassedNum,
		"failedCaseNum":    failedNum,
		"passedCaseNum":    passedNum,
		"casesList":        []map[string]interface{}{},
	}

	domains := []string{"storage", "compute", "network", "other"}

	cases := make(map[string]map[string]interface{})
	for _, domain := range domains {
		cases[domain] = map[string]interface{}{
			"domain":            domain,
			"totalCasesNum":     0,
			"notPassedCasesNum": 0,
			"failedCasesNum":    0,
			"passedCasesNum":    0,
			"failed":            []map[string]interface{}{},
			"passed":            []map[string]interface{}{},
			"notPassed":         []map[string]interface{}{},
		}
	}

	for _, record := range records {
		domainInfo, e1 := models.QueryScriptByName(record.CaseName, "domain")
		categoryInfo, e2 := models.QueryScriptByName(record.CaseName, "category")
		chNameInfo, e3 := models.QueryScriptByName(record.CaseName, "ch_name")

		// 如果任何查询返回错误，则跳过当前记录
		if e1 != nil || e2 != nil || e3 != nil {
			log.Printf("Error querying scripts for case %s: domain=%v, category=%v, ch_name=%v", record.CaseName, e1, e2, e3)
			continue
		}

		domain, ok := domainInfo.(string)
		if !ok {
			log.Printf("Error: domainInfo is not a string for case %s", record.CaseName)
			continue
		}

		caseDetail := map[string]interface{}{
			"caseName":         chNameInfo.(string),
			"category":         categoryInfo.(string),
			"healthyNodeNum":   record.HealthyNum,
			"criticalNodeNum":  record.CriticalNum,
			"unhealthyNodeNum": record.UnhealthyNum,
			"unknownNodeNum":   record.UnknownNum,
			"timeoutNodeNum":   record.TimeoutNum,
			"totalNodeNum":     record.TotalNum,
		}

		cases[domain]["totalCasesNum"] = cases[domain]["totalCasesNum"].(int) + 1
		switch record.Result {
		case "passed":
			cases[domain]["passedCasesNum"] = cases[domain]["passedCasesNum"].(int) + 1
			cases[domain]["passed"] = append(cases[domain]["passed"].([]map[string]interface{}), caseDetail)
		case "not passed":
			cases[domain]["notPassedCasesNum"] = cases[domain]["notPassedCasesNum"].(int) + 1
			cases[domain]["notPassed"] = append(cases[domain]["notPassed"].([]map[string]interface{}), caseDetail)
		case "failed":
			cases[domain]["failedCasesNum"] = cases[domain]["failedCasesNum"].(int) + 1
			cases[domain]["failed"] = append(cases[domain]["failed"].([]map[string]interface{}), caseDetail)
		}
	}

	for _, domain := range domains {
		if value, exists := cases[domain]; exists {
			result["casesList"] = append(result["casesList"].([]map[string]interface{}), value)
		}
	}

	return result, nil
}

func getScriptDetailInfo(jobId, caseChName string) (interface{}, error) {
	jobName, err := models.QueryJobInfoById(jobId, "job_name")
	if err != nil {
		return nil, err
	}
	createTime, err := models.QueryJobInfoById(jobId, "create_time")
	if err != nil {
		return nil, err
	}

	scriptName, err := models.QueryScriptNameByChName(caseChName)
	if err != nil {
		return nil, err
	}
	testInfo, err := models.GetTestRecordByCase(jobId, scriptName)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"jobName":          jobName,
		"createTime":       createTime,
		"timeCost":         testInfo.TimeCost,
		"result":           testInfo.Result,
		"healthyNodeNum":   testInfo.HealthyNum,
		"unhealthyNodeNum": testInfo.UnhealthyNum,
		"criticalNodeNum":  testInfo.CriticalNum,
		"unknownNodeNum":   testInfo.UnknownNum,
		"timeoutNodeNum":   testInfo.TimeoutNum,
		"totalNodeNum":     testInfo.TotalNum,
		"logFileName":      scriptName,
		"log":              "",
	}

	// 获取日志文件
	logPath, err := beego.AppConfig.String("log_path")
	if err != nil {
		return nil, fmt.Errorf("failed to get log_path: %w", err)
	}

	filePath := filepath.Join(logPath, jobId, scriptName+".log")
	fileData, readErr := os.ReadFile(filePath)

	if readErr == nil {
		result["log"] = parseLogFile(string(fileData))
	} else if os.IsNotExist(readErr) {
		result["log"] = base64.StdEncoding.EncodeToString([]byte("no log exists"))
	} else {
		return nil, readErr
	}

	return result, nil
}

// Node 结构体定义节点信息
type Node struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

// CheckResult 结构体定义一个健康检查记录，包含节点列表和是否为多节点
type CheckResult struct {
	Status      string `json:"status"`
	Detail      string `json:"detail,omitempty"`
	Nodes       []Node `json:"nodes"`
	IsMultiNode bool   `json:"isMultiNode"`
}

// parseLogFile 解析文本内容为 CheckResult 结构体切片
func parseLogFile(text string) []CheckResult {
	// 分割文本为行
	lines := strings.Split(strings.TrimSpace(text), "\n")

	// 初始化一个空的 CheckResult 切片
	var results []CheckResult

	// 遍历每一行
	for _, line := range lines {
		// 按第一个冒号分割行
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue // 跳过没有冒号的行
		}

		// 提取状态和节点部分（冒号前内容）
		statusAndNodes := strings.Fields(parts[0]) // 冒号前的部分按空格分割
		if len(statusAndNodes) < 2 {
			continue // 跳过不合法的行
		}

		status := strings.Trim(statusAndNodes[0], "[]") // 去掉状态两侧的中括号
		nodes := statusAndNodes[1]

		// 提取冒号后的内容（message）
		message := strings.TrimSpace(parts[1]) // 冒号后的部分

		result := &CheckResult{
			Status:      status,
			Detail:      message,
			Nodes:       parseMultipleNodes(nodes),
			IsMultiNode: false, // 初始化为 false
		}

		// 判断是否为多节点记录
		if len(result.Nodes) > 1 {
			result.IsMultiNode = true
		}

		results = append(results, *result)
	}

	return results
}

func parseMultipleNodes(input string) []Node {
	// 用于存储解析后的节点
	var nodes []Node

	// 按 `&` 分割多个节点
	nodeParts := strings.Split(input, "&")

	// 遍历每个节点部分
	for _, part := range nodeParts {
		// 分离 IP 和 Name
		ipNameParts := strings.Split(part, "(")
		if len(ipNameParts) != 2 {
			continue // 跳过不合法的格式
		}

		ip := ipNameParts[0]                            // 获取 IP
		name := strings.TrimRight(ipNameParts[1], "):") // 获取 Name，去掉右侧的括号

		// 添加到节点列表
		nodes = append(nodes, Node{
			IP:   ip,
			Name: name,
		})
	}

	return nodes
}
