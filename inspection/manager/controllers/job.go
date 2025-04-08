package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/api"
	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/job"
)

type JobController struct {
	beego.Controller
}

// Post ...
// @Summary     创建作业任务
// @Description 创建新的作业任务，需提供作业详情以及相关请求头信息
// @Tags        作业管理
// @Accept      json
// @Produce     json
// @Param       x-username header  string true "请求者的用户名，需进行 URL 编码"
// @Param       x-cluster  header  string true "作业执行的集群名称"
// @Param       body       body    job.CreateJobRequest true "包含作业详细信息的 JSON 数据"
// @Success     200 {object} api.CommonResponse "成功响应，返回创建的作业 ID"
// @Failure     400 {object} api.CommonResponse "参数无效或缺少必要字段"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/create_job [post]
func (c *JobController) Post() {
	// 从请求头中获取 x-username
	xUserName := c.Ctx.Input.Header("x-username")
	clusterName := c.Ctx.Input.Header("x-cluster")

	// 解码用户名
	userName, err := url.QueryUnescape(xUserName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse("Failed to decode username")
		c.ServeJSON()
		return
	}

	// 解析 JSON 请求体
	var req job.CreateJobRequest
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = api.ParamErrResponse("Invalid JSON payload")
		c.ServeJSON()
		return
	}

	// 检查必需字段
	if xUserName == "" || clusterName == "" {
		c.Data["json"] = api.ParamErrResponse("x-username and x-cluster are required")
		c.ServeJSON()
		return
	}
	if req.JobName == "" || req.TemplateName == "" || req.Mode == "" || req.Resource == "" {
		c.Data["json"] = api.ParamErrResponse("JobName, Template, Mode, Resource are required")
		c.ServeJSON()
		return
	}

	// 生成作业 ID
	currentTime := time.Now()
	jobId := fmt.Sprintf("%s-%s", currentTime.Format("20060102"), generateRandomString(6))
	ipListStr := strings.Join(req.IpList, ",")

	templateID, err := models.GetTemplateIdByName(req.TemplateName, clusterName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(fmt.Sprintf("GetTemplateIdByName error: %v", err))
		c.ServeJSON()
		return
	}

	// 准备作业信息
	jobInfo := models.JobInfo{
		UserName:     userName,
		ClusterName:  clusterName,
		TemplateName: req.TemplateName,
		TemplateID:   templateID,
		JobName:      req.JobName,
		JobId:        jobId,
		Status:       "creating",
		IsCron:       req.IsCron,
		Mode:         req.Mode,
		Resource:     req.Resource,
		IpList:       ipListStr,
		BaseIp:       req.BaseIP,
		CreateTime:   currentTime.Format("2006-01-02 15:04:05"),
	}

	if err = models.AddJobInfo(&jobInfo); err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	err = job.CreateJob(jobInfo)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(fmt.Sprintf("Failed to create task: %v", err))
	} else {
		c.Data["json"] = api.SuccessResponse(jobInfo.JobId)
	}

	c.ServeJSON()
}

// Retry ...
// @Summary     重试作业
// @Description 根据作业 ID 重试一个作业，从数据库中获取作业详情，并使用相同的参数创建新作业
// @Tags        作业管理
// @Accept      json
// @Produce     json
// @Param       x-username header  string true "发起请求的用户名（需进行 URL 编码）"
// @Param       x-cluster  header  string true "与作业关联的集群名称"
// @Param       jobId      body    object true "包含需要重试的作业 ID 的 JSON 对象" Example: {"jobId": "example-job-id"}
// @Success     200 {object} api.CommonResponse "重试成功的响应，包含作业详情"
// @Failure     400 {object} api.CommonResponse "x-username 和 x-cluster 为必填项"
// @Failure     500 {object} api.CommonResponse "获取作业 ID 失败"
// @Router      /inspection/api/v1/job/retry [post]
func (c *JobController) Retry() {
	// 从请求头中获取 x-username
	xUserName := c.Ctx.Input.Header("x-username")
	clusterName := c.Ctx.Input.Header("x-cluster")

	if xUserName == "" || clusterName == "" {
		c.Data["json"] = api.ParamErrResponse("x-username and x-cluster are required")
		c.ServeJSON()
		return
	}

	// 解码用户名
	userName, err := url.QueryUnescape(xUserName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse("Failed to decode username")
		c.ServeJSON()
		return
	}

	var req struct {
		JobId string `json:"jobId"`
	}

	// 解析请求体中的 JSON
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = api.ParamErrResponse("Failed to get job id")
		c.ServeJSON()
		return
	}

	var jobInfo *models.JobInfo

	if jobInfo, err = models.GetJobInfo(req.JobId); err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	jobParam := job.CreateJobRequest{
		JobName:      jobInfo.JobName,
		TemplateName: jobInfo.TemplateName,
		IsCron:       false,
		Mode:         jobInfo.Mode,
		IpList:       strings.Split(jobInfo.IpList, ","),
		BaseIP:       jobInfo.BaseIp,
	}

	job.CallCreateUrl(jobParam, userName, clusterName)

	c.Data["json"] = api.SuccessResponse(fmt.Sprintf("Retry job %s successfully.", jobInfo.JobId))
	c.ServeJSON()
}

// GetDetail ...
// @Summary     获取作业详情
// @Description 根据作业 ID 获取作业的详细信息
// @Tags        作业管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称"
// @Param       jobID     query   string true "需要获取详情的作业 ID"
// @Success     200 {object} api.CommonResponse "成功返回作业的详细信息"
// @Failure     400 {object} api.CommonResponse "x-cluster 或 jobID 为必填项"
// @Failure     500 {object} api.CommonResponse "获取作业详情失败"
// @Router      /inspection/api/v1/job/detail [get]
func (c *JobController) GetDetail() {
	clusterName := c.Ctx.Input.Header("x-cluster")

	jobID := c.GetString("jobID")

	if clusterName == "" || jobID == "" {
		c.Data["json"] = api.ParamErrResponse("x-cluster and jobID are required")
		c.ServeJSON()
		return
	}

	var jobInfo *models.JobInfo

	jobInfo, err := models.GetJobInfo(jobID)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	jobDetail := job.CreateJobRequest{
		JobName:      jobInfo.JobName,
		TemplateName: jobInfo.TemplateName,
		IsCron:       jobInfo.IsCron,
		Mode:         jobInfo.Mode,
		IpList:       strings.Split(jobInfo.IpList, ","),
		BaseIP:       jobInfo.BaseIp,
	}

	c.Data["json"] = api.SuccessResponse(jobDetail)
	c.ServeJSON()
}

// GetAll ...
// @Summary     获取所有作业
// @Description 获取作业信息，根据查询条件和分页参数返回作业列表
// @Tags        作业管理
// @Accept      json
// @Produce     json
// @Param       pageID            query   int    false "分页页码，默认值为 1"
// @Param       pageSize          query   int    false "每页记录数，默认值为 10"
// @Param       cluster_name      header  string false "根据集群名称过滤作业记录"
// @Param       jobName           query   string false "根据作业名称过滤"
// @Param       mode              query   string false "根据模式过滤"
// @Param       status            query   string false "根据作业状态过滤"
// @Param       templateName      query   string false "根据模板名称过滤"
// @Param       finishedTimeRange query   string false "根据完成时间范围过滤"
// @Success     200 {object} api.CommonResponse "成功返回作业记录列表的响应"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/test_result_info [get]
func (c *JobController) GetAll() {
	var pageID int64 = 1
	var pageSize int64 = 10
	var sortBy = "create_time"
	var sortOrder = "desc"
	var sort string
	var query = make(map[string]string)

	if v, err := c.GetInt64("currentPage"); err == nil {
		pageID = v
	}
	if v, err := c.GetInt64("pageSize"); err == nil {
		pageSize = v
	}

	if v := c.Ctx.Input.Header("x-cluster"); v != "" {
		query["cluster_name"] = v
	}
	if v := c.GetString("jobName"); v != "" {
		query["job_name"] = v
	}
	if v := c.GetString("mode"); v != "" {
		query["mode"] = v
	}
	if v := c.GetString("status"); v != "" {
		query["status"] = v
	}
	if v := c.GetString("templateName"); v != "" {
		query["template_name"] = v
	}
	if v := c.GetString("createTimeRange"); v != "" {
		query["createTimeRange"] = v
	}
	if v := c.GetString("finishedTimeRange"); v != "" {
		query["finishedTimeRange"] = v
	}
	if v := c.GetString("sortBy"); v != "" {
		if v == "finishedTime" {
			sortBy = "finish_time"
		}
	}
	if v := c.GetString("sortOrder"); v != "" {
		sortOrder = v
	}

	if sortOrder == "desc" {
		sort = "-" + sortBy
	} else {
		sort = sortBy
	}

	records, totalNum, err := models.GetJobRecords(query, pageID, pageSize, sort)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	result, err := processJobRecords(records, pageID, pageSize, totalNum)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(result)
	c.ServeJSON()
}

func generateRandomString(length int) string {
	chars := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func processJobRecords(records []models.JobInfo, pageID int64, pageSize int64, totalNum int64) (interface{}, error) {
	result := map[string]interface{}{
		"result": []map[string]interface{}{},
		"page": map[string]interface{}{
			"pageSize":    pageSize,
			"currentPage": pageID,
			"total":       totalNum,
		},
	}
	for _, record := range records {
		// [已完成用例数量, 总用例数量]
		finishTotal := []int{0, 0}
		finished := 0
		total := 0

		testInfo, err := models.GetTestRecords(record.JobId)
		if err == nil {
			finished = len(testInfo)
		}

		// 查询模板下的用例数量
		scriptNames, err := models.QueryScriptsByTemplateID(record.TemplateID)
		if err == nil {
			total = len(scriptNames)
		}

		if total < finished {
			total = finished
		}

		finishTotal[0] = finished
		finishTotal[1] = total

		result["result"] = append(result["result"].([]map[string]interface{}), map[string]interface{}{
			"jobId":        record.JobId,
			"jobName":      record.JobName,
			"userName":     record.UserName,
			"mode":         record.Mode,
			"status":       record.Status,
			"cluster":      record.ClusterName,
			"template":     record.TemplateName,
			"isCron":       record.IsCron,
			"createTime":   record.CreateTime,
			"finishedTime": record.FinishTime,
			"timeCost":     getJobTimeCost(record.JobId),
			"finishTotal":  finishTotal,
		})
	}

	return result, nil
}

func getJobTimeCost(jobId string) string {
	jobCreateTime, err := models.QueryJobInfoById(jobId, "create_time")
	if err != nil {
		return ""
	}
	jobFinishTime, err := models.QueryJobInfoById(jobId, "finish_time")
	if err != nil {
		return ""
	}
	if jobFinishTime == "" {
		return ""
	}

	// 将时间字符串解析为 time.Time 对象
	layout := "2006-01-02 15:04:05"
	startTime, err := time.Parse(layout, jobCreateTime.(string))
	if err != nil {
		fmt.Println("解析开始时间时出错:", err)
		return ""
	}
	finishTime, err := time.Parse(layout, jobFinishTime.(string))
	if err != nil {
		fmt.Println("解析结束时间时出错:", err)
		return ""
	}

	// 计算时间差
	duration := finishTime.Sub(startTime)
	// 获取时、分、秒
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	// 格式化输出
	formattedDuration := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	return formattedDuration
}
