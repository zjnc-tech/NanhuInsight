package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/api"
	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/cronjob"
	"infrahi/backend/inspection-manager/pkg/templates"
)

type TemplateController struct {
	beego.Controller
}

// Get ...
// @Summary     获取所有模板
// @Description 根据查询条件、分页参数和排序选项返回模板记录
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       pageID     query   int    false "分页页码，默认值为 1"
// @Param       pageSize   query   int    false "每页记录数，默认值为 10"
// @Param       x-cluster  header  string false "根据集群名称过滤模板记录"
// @Param       mode       query   string false "根据模式过滤"
// @Param       cron       query   string false "根据定时任务状态过滤"
// @Param       sortBy     query   string false "排序字段，默认值为 'create_time'"
// @Param       sortOrder  query   string false "排序顺序，可选值为 'asc' 或 'desc'，默认值为 'desc'"
// @Success     200 {object} api.CommonResponse "成功返回模板记录列表的响应"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/get_template_info [get]
func (c *TemplateController) Get() {
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
	if v := c.GetString("mode"); v != "" {
		query["mode"] = v
	}
	if v := c.GetString("cron"); v != "" {
		query["cron"] = v
	}
	if v := c.GetString("sortBy"); v != "" {
		if v == "modifyTime" {
			sortBy = "modify_time"
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

	records, totalNum, err := models.GetTemplateRecords(query, pageID, pageSize, sort)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	result, err := processTemplateRecords(records, pageID, pageSize, totalNum)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(result)
	c.ServeJSON()
}

// GetAll ...
// @Summary     根据模式获取模板
// @Description 根据集群名称、模式和资源名称查询模板列表
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于过滤模板（必填项）"
// @Param       mode      query   string true "模式，用于过滤模板（必填项）"
// @Param       resource query   string false "资源名称，用于过滤模板（可选项）"
// @Success     200 {object} api.CommonResponse "成功返回模板名称列表的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/get_all_templates [get]
func (c *TemplateController) GetAll() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	mode := c.GetString("mode")
	resource := c.GetString("resource")

	if clusterName == "" || mode == "" || resource == "" {
		c.Data["json"] = api.ParamErrResponse("Param cluster, resource and mode are required")
		c.ServeJSON()
		return
	}

	data, err := models.QueryTemplateByMode(clusterName, mode, resource)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(data)
	c.ServeJSON()
}

// GetOne ...
// @Summary     获取模板详情
// @Description 根据集群名称和模板名称查询模板的详细信息
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，用于过滤模板（必填项）"
// @Param       template  query   string true "模板名称（必填项）"
// @Success     200 {object} api.CommonResponse "成功返回模板详情的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/get_template_detail [get]
func (c *TemplateController) GetOne() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	templateName := c.GetString("template")

	if clusterName == "" || templateName == "" {
		c.Data["json"] = api.ParamErrResponse("cluster and template are required")
		c.ServeJSON()
		return
	}

	detail, err := models.QueryDetailOfTemplate(templateName, clusterName)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(detail)
	c.ServeJSON()
}

// Post ...
// @Summary     创建模板
// @Description 创建新的模板，需提供模板详细信息以及相关请求头信息
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       x-username header  string true "请求者的用户名，需进行 URL 编码"
// @Param       x-cluster  header  string true "模板创建所在的集群名称"
// @Param       body       body    templates.CreateRequest true "包含模板详细信息的 JSON 数据"
// @Success     200 {object} api.CommonResponse "成功返回创建模板的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/create_template [post]
func (c *TemplateController) Post() {
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

	// 解析 JSON 请求体
	var req templates.CreateRequest
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = api.ParamErrResponse("Invalid JSON payload")
		c.ServeJSON()
		return
	}

	if req.TemplateName == "" || req.Mode == "" {
		c.Data["json"] = api.ParamErrResponse("template and mode are required")
		c.ServeJSON()
		return
	}

	templateId, err := models.GetTemplateIdByName(req.TemplateName, clusterName)
	if err == nil && templateId != -1 {
		c.Data["json"] = api.ParamErrResponse(fmt.Sprintf("%s集群中已存在名为%s的模板，请更换模板名.",
			clusterName, req.TemplateName))
		c.ServeJSON()
		return
	}

	newTemplate := &models.Template{
		TemplateName: req.TemplateName,
		ClusterName:  clusterName,
		Description:  req.Description,
		Mode:         req.Mode,
		Resource:     req.Resource,
		CreateUser:   userName,
		CreateTime:   time.Now().Format("2006-01-02 15:04:05"),
	}

	if _, err = models.AddTemplate(newTemplate, req.ScriptList); err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	if req.NeedCron {
		if err = cronjob.CreateCronJob(req, userName, clusterName); err != nil {
			log.Printf("failed to create cron job: %v", err)
			deleteErr := models.DeleteTemplate(req.TemplateName, clusterName)
			if deleteErr != nil {
				log.Printf("failed to delete cron job:%v", deleteErr)
			}
			c.Data["json"] = api.ParamErrResponse(err.Error())
			c.ServeJSON()
		}
	}

	c.Data["json"] = api.SuccessResponse(fmt.Sprintf("Template created successfully."))
	c.ServeJSON()
}

// Patch ...
// @Summary     修改模板
// @Description 修改模板的基本信息或定时任务配置
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       x-username header  string true "请求者的用户名，需进行 URL 编码"
// @Param       x-cluster  header  string true "模板所在集群的名称"
// @Param       body       body    templates.CreateRequest true "包含模板修改信息的 JSON 数据"
// @Success     200 {object} api.CommonResponse "成功返回模板修改的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/modify_template [patch]
func (c *TemplateController) Patch() {
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

	// 解析 JSON 请求体
	var req templates.CreateRequest
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = api.ParamErrResponse("Invalid JSON payload")
		c.ServeJSON()
		return
	}

	if req.TemplateName == "" || req.Mode == "" {
		c.Data["json"] = api.ParamErrResponse("template and mode are required")
		c.ServeJSON()
		return
	}

	templateId, err := models.GetTemplateIdByName(req.TemplateName, clusterName)
	if err == nil && templateId == -1 {
		c.Data["json"] = api.ParamErrResponse(fmt.Sprintf("%s集群中不存在名为%s的模板.",
			clusterName, req.TemplateName))
		c.ServeJSON()
		return
	}

	basicInfo := &models.TemplateBasicInfo{
		Mode:        req.Mode,
		Resource:    req.Resource,
		Description: req.Description,
		ModifyTime:  time.Now().Format("2006-01-02 15:04:05"),
		ModifyUser:  userName,
	}

	if err = models.UpdateTemplateBasicInfo(req.TemplateName, clusterName, basicInfo, req.ScriptList); err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	err = cronjob.DeleteCronJob(req.TemplateName, clusterName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
	}

	if req.NeedCron {
		err = cronjob.CreateCronJob(req, userName, clusterName)
		if err != nil {
			c.Data["json"] = api.ParamErrResponse(err.Error())
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = api.SuccessResponse(fmt.Sprintf("Template modified successfully."))
	c.ServeJSON()
}

// Delete ...
// @Summary     删除模板
// @Description 删除指定集群中的模板
// @Tags        模板管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "模板所在集群的名称"
// @Param       template  query   string true "要删除的模板名称"
// @Success     200 {object} api.CommonResponse "成功返回模板删除的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/delete_template [delete]
func (c *TemplateController) Delete() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	templateName := c.GetString("template")

	if clusterName == "" || templateName == "" {
		c.Data["json"] = api.ParamErrResponse("cluster and template are required")
		c.ServeJSON()
		return
	}

	if err := cronjob.RemoveScheduleJob(templateName, clusterName); err != nil {
		log.Printf("failed to remove cron job: %v", err)
	}

	err := models.DeleteTemplate(templateName, clusterName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse("Template deleted successfully.")
	c.ServeJSON()
}

func processTemplateRecords(records []models.Template, pageID int64, pageSize int64,
	totalNum int64) (interface{}, error) {
	result := map[string]interface{}{
		"result": []map[string]interface{}{},
		"page": map[string]interface{}{
			"pageSize":    pageSize,
			"currentPage": pageID,
			"total":       totalNum,
		},
	}
	for _, record := range records {
		cronTime := ""
		if record.CronMinute != "" {
			if record.CronFrequency == "minutely" {
				cronTime = fmt.Sprintf("%02s", record.CronMinute)
			} else {
				cronTime = fmt.Sprintf("%02s:%02s", record.CronHour, record.CronMinute)
			}
			//if record.CronHour != "" {
			//	cronTime = fmt.Sprintf("%02s:%02s", record.CronHour, record.CronMinute)
			//} else {
			//	cronTime = fmt.Sprintf("%02s", record.CronMinute)
			//}
		}

		result["result"] = append(result["result"].([]map[string]interface{}), map[string]interface{}{
			"template":    record.TemplateName,
			"description": record.Description,
			"mode":        record.Mode,
			"needCron":    record.CronSwitch,
			"frequency":   record.CronFrequency,
			"cronTime":    cronTime,
			"dayOfWeek":   record.DayOfWeek,
			"dayOfMonth":  record.DayOfMonth,
			"lastJobTime": record.LastJobTime,
			"lastJobId":   record.LastJobID,
			"createUser":  record.CreateUser,
			"createTime":  record.CreateTime,
			"modifyUser":  record.ModifyUser,
			"modifyTime":  record.ModifyTime,
		})
	}

	return result, nil
}
