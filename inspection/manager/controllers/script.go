package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/api"
	"infrahi/backend/inspection-manager/models"
)

type ScriptController struct {
	beego.Controller
}

type SetConfigRequest struct {
	ScriptName   string                 `json:"scriptName"`
	TemplateName string                 `json:"templateName"`
	Params       map[string]interface{} `json:"params"`
}

// Get ...
// @Summary     获取脚本
// @Description 根据指定的模式检索脚本列表
// @Tags        脚本管理
// @Accept      json
// @Produce     json
// @Param       x-cluster header  string true "集群名称，指定脚本所在的集群"
// @Param       mode      query   string true "过滤检查项的模式，根据指定的模式筛选"
// @Param       resource  query   string true "资源名称，根据指定的资源筛选检查项"
// @Success     200 {object} api.CommonResponse "成功返回包含脚本列表的响应"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/get_all_cases [get]
func (c *ScriptController) Get() {
	clusterName := c.Ctx.Input.Header("x-cluster")
	resource := c.GetString("resource")
	mode := c.GetString("mode")

	if mode == "" || resource == "" {
		c.Data["json"] = api.ParamErrResponse("param mode and resource are required")
		c.ServeJSON()
		return
	}

	result, err := models.GetScriptsByMode(mode, clusterName, resource)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(result)
	c.ServeJSON()
}

// GetConfig ...
// @Summary     获取脚本配置
// @Description 根据指定的多个脚本名称获取其配置信息
// @Tags        脚本管理
// @Accept      json
// @Produce     json
// @Param       scriptNames query   string true "脚本名称列表，多个脚本名称用逗号分隔，例如 'script1,script2'"
// @Success     200 {object} map[string][]models.ParamResponse "成功返回包含脚本配置的响应，每个脚本名称对应一个参数列表"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/script/get_config [get]
func (c *ScriptController) GetConfig() {
	scriptNamesStr := c.GetString("scriptNames")
	if scriptNamesStr == "" {
		c.Data["json"] = api.ParamErrResponse("scriptNames is required")
		c.ServeJSON()
		return
	}

	// 将逗号分隔的字符串转成数组
	scriptNames := strings.Split(scriptNamesStr, ",")

	result, err := models.GetScriptConfigs(scriptNames)
	if err != nil {
		c.Data["json"] = api.InternalErrResponse(err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = api.SuccessResponse(result)
	c.ServeJSON()
}

// SetConfig ...
// @Summary     设置检查项的参数值
// @Description 对指定模板中的指定检查项设置参数的名称与值
// @Tags        脚本管理
// @Accept      json
// @Produce     json
// @Param       x-cluster  header  string true "指定集群名称"
// @Param       body       body    controllers.SetConfigRequest true "包含检查项参数名称和值的 JSON 数据"
// @Success     200 {object} api.CommonResponse "成功设置参数"
// @Failure     400 {object} api.CommonResponse "请求错误，例如缺少必要参数或参数无效"
// @Failure     500 {object} api.CommonResponse "服务器内部错误"
// @Router      /inspection/api/v1/script/set_config [get]
func (c *ScriptController) SetConfig() {
	clusterName := c.Ctx.Input.Header("x-cluster")

	var req SetConfigRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = api.ParamErrResponse("Invalid JSON payload")
		c.ServeJSON()
		return
	}

	templateId, err := models.GetTemplateIdByName(req.TemplateName, clusterName)
	if err != nil {
		c.Data["json"] = api.ParamErrResponse(fmt.Sprintf("GetTemplateIdByName error: %v", err))
		c.ServeJSON()
		return
	}
	if templateId == -1 {
		c.Data["json"] = api.ParamErrResponse(fmt.Sprintf("%s集群中不存在名为%s的模板.",
			clusterName, req.TemplateName))
		c.ServeJSON()
		return
	}

	for key, value := range req.Params {
		// 创建 ScriptConfig 实例
		scriptConfig := models.ScriptConfig{
			ScriptName:  req.ScriptName,
			TemplateID:  templateId,
			ConfigName:  key,
			ConfigValue: fmt.Sprintf("%v", value), // 将 value 转换为字符串
		}

		// 插入到数据库
		if err = models.AddScriptConfig(&scriptConfig); err != nil {
			log.Printf("Error inserting data for key %s: %v", key, err)
		} else {
			log.Printf("Inserted config: %s = %v\n", key, value)
		}
	}

	c.Data["json"] = api.SuccessResponse("Config set.")
	c.ServeJSON()
}
