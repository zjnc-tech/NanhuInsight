package models

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/beego/beego/v2/client/orm"

	"infrahi/backend/inspection-manager/pkg/templates"
)

type Template struct {
	TemplateID    int64     `json:"template_id" orm:"column(template_id);pk;auto"`
	TemplateName  string    `json:"template_name" orm:"size(128)"`
	ClusterName   string    `json:"cluster_name" orm:"size(128)"`
	Description   string    `json:"description" orm:"type(text)"`
	Mode          string    `json:"mode" orm:"size(20);default(regular)"`
	Resource      string    `json:"resource" orm:"size(128)"`
	LastJobTime   string    `json:"last_job_time" orm:"null"`
	LastJobID     string    `json:"last_job_id" orm:"column(last_job_id);size(128)"`
	CronSwitch    bool      `json:"cron_switch" orm:"default(false)"`
	CronExpr      string    `json:"cron_expr" orm:"column(cron_expr);size(255)"`
	CronFrequency string    `json:"cron_frequency" orm:"size(128)"`
	CronHour      string    `json:"cron_hour" orm:"size(128)"`
	CronMinute    string    `json:"cron_minute" orm:"size(128)"`
	DayOfWeek     string    `json:"day_of_week" orm:"size(128)"`
	DayOfMonth    string    `json:"day_of_month" orm:"size(128)"`
	JobName       string    `json:"job_name" orm:"size(128)"`
	IpList        string    `json:"ip_list" orm:"type(text);null"`
	BaseIp        string    `json:"base_ip" orm:"size(128);null"`
	CreateUser    string    `json:"create_user" orm:"size(128)"`
	CreateTime    string    `json:"create_time" orm:"auto_now_add"`
	ModifyUser    string    `json:"modify_user" orm:"size(128)"`
	ModifyTime    string    `json:"modify_time" orm:"null"`
	Scripts       []*Script `json:"scripts" orm:"rel(m2m)"`
}

type TemplateBasicInfo struct {
	Mode        string    `json:"mode"`
	Resource    string    `json:"resource"`
	Description string    `json:"description"`
	ModifyUser  string    `json:"modify_user"`
	ModifyTime  string    `json:"modify_time"`
	Scripts     []*Script `json:"scripts"`
}

type TemplateCronInfo struct {
	CronSwitch    bool   `json:"cron_switch"`
	CronExpr      string `json:"cron_expr"`
	CronFrequency string `json:"cron_frequency"`
	CronHour      string `json:"cron_hour"`
	CronMinute    string `json:"cron_minute"`
	DayOfWeek     string `json:"day_of_week"`
	DayOfMonth    string `json:"day_of_month"`
	JobName       string `json:"job_name"`
	IpList        string `json:"ip_list"`
	BaseIp        string `json:"base_ip"`
}

type TemplateScripts struct {
	Template *Template `orm:"rel(fk)"`
	Script   *Script   `orm:"rel(fk)"`
}

func InitTemplate() {
	orm.RegisterModel(new(Template))
}

func AddTemplate(template *Template, scriptParams []templates.ScriptParam) (id int64, err error) {
	o := orm.NewOrm()

	templateId, err := o.Insert(template)
	if err != nil {
		return 0, fmt.Errorf("error inserting template: %w", err)
	}

	for _, scriptParam := range scriptParams {
		err = AssociateScript(template, scriptParam.ScriptName)
		if err != nil {
			return 0, err
		}

		for key, value := range scriptParam.Params {
			// 创建 ScriptConfig 实例
			scriptConfig := ScriptConfig{
				ScriptName:  scriptParam.ScriptName,
				TemplateID:  templateId,
				ConfigName:  key,
				ConfigValue: fmt.Sprintf("%v", value), // 将 value 转换为字符串
			}

			// 插入到数据库
			if err = AddScriptConfig(&scriptConfig); err != nil {
				log.Printf("Error inserting data for key %s: %v", key, err)
			} else {
				log.Printf("Inserted config: %s = %v\n", key, value)
			}
		}
	}
	return templateId, nil
}

func DeleteTemplate(templateName, clusterName string) (err error) {
	o := orm.NewOrm()

	// 获取模板以确保其存在
	var t Template
	err = o.QueryTable("template").Filter("template_name", templateName).
		Filter("cluster_name", clusterName).One(&t)
	if err != nil {
		return fmt.Errorf("error finding template: %w", err)
	}

	// 删除与模板的关联脚本
	_, err = o.QueryM2M(&t, "Scripts").Clear()
	if err != nil {
		return fmt.Errorf("error clearing associations: %w", err)
	}

	// 删除相关配置
	_, err = o.QueryTable("script_config").Filter("template_id", t.TemplateID).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete script configs for template_id %d: %v", t.TemplateID, err)
	}

	// 删除模板
	_, err = o.Delete(&t)
	if err != nil {
		return fmt.Errorf("error deleting template: %w", err)
	}

	return nil
}

func GetCronTemplates() (templates []Template, err error) {
	o := orm.NewOrm()

	_, err = o.QueryTable("template").Filter("cron_switch", true).All(&templates)
	if err != nil {
		return nil, fmt.Errorf("error finding templates: %w", err)
	}

	return templates, nil
}

func GetTemplateRecords(query map[string]string, pageID int64, pageSize int64,
	sort string) (templates []Template, totalNum int64, err error) {
	o := orm.NewOrm()

	qs := o.QueryTable("template")

	// 可按需添加查询条件
	for key, value := range query {
		if key == "cron" {
			if value == "on" {
				qs = qs.Filter("cron_switch", true)
			} else {
				qs = qs.Filter("cron_switch", false)
			}
		} else {
			qs = qs.Filter(key, value)
		}
	}

	qs = qs.OrderBy(sort)

	// 统计总记录数（用于计算总页数）
	totalNum, err = qs.Count()
	if err != nil {
		return nil, 0, err
	}

	_, err = qs.Limit(pageSize, (pageID-1)*pageSize).All(&templates)
	if err != nil {
		return nil, 0, err
	}

	return templates, totalNum, nil
}

func GetTemplateIdByName(templateName, clusterName string) (int64, error) {
	o := orm.NewOrm()
	var t Template

	qs := o.QueryTable("template").Filter("template_name", templateName).
		Filter("cluster_name", clusterName)
	err := qs.One(&t)
	if errors.Is(orm.ErrNoRows, err) {
		return -1, nil
	} else if err != nil {
		log.Println("Error querying data:", err)
		return 0, err
	} else {
		return t.TemplateID, nil
	}
}

func UpdateTemplateJobInfo(templateId int64, jobId, lastJobTime string) (err error) {
	o := orm.NewOrm()
	var template Template

	qs := o.QueryTable("template").Filter("template_id", templateId)
	if err = qs.One(&template); err == nil {
		template.LastJobID = jobId
		template.LastJobTime = lastJobTime
		if _, err = o.Update(&template); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func UpdateTemplateBasicInfo(templateName string, clusterName string, basicInfo *TemplateBasicInfo,
	scriptParams []templates.ScriptParam) (err error) {
	o := orm.NewOrm()

	var template Template

	qs := o.QueryTable("template").Filter("template_name", templateName).
		Filter("cluster_name", clusterName)
	if err = qs.One(&template); err == nil {
		template.Mode = basicInfo.Mode
		template.Resource = basicInfo.Resource
		template.Description = basicInfo.Description
		template.ModifyUser = basicInfo.ModifyUser
		template.ModifyTime = basicInfo.ModifyTime

		// 获取模板的 ID
		templateID := template.TemplateID

		// 删除现有关联
		_, err = o.QueryTable("template_scripts").Filter("template_id", templateID).Delete()
		if err != nil {
			return fmt.Errorf("error deleting existing associations: %w", err)
		}

		// 删除相关配置
		_, err := o.QueryTable("script_config").Filter("template_id", templateID).Delete()
		if err != nil {
			log.Printf("Failed to delete records for template_id %d: %v", templateID, err)
			return fmt.Errorf("failed to delete records for template_id %d: %v", templateID, err)
		}

		// 添加新的关联
		for _, scriptParam := range scriptParams {
			err = AssociateScript(&template, scriptParam.ScriptName)
			if err != nil {
				return err
			}

			for key, value := range scriptParam.Params {
				// 创建 ScriptConfig 实例
				scriptConfig := ScriptConfig{
					ScriptName:  scriptParam.ScriptName,
					TemplateID:  templateID,
					ConfigName:  key,
					ConfigValue: fmt.Sprintf("%v", value), // 将 value 转换为字符串
				}

				// 插入到数据库
				if err = AddScriptConfig(&scriptConfig); err != nil {
					log.Printf("Error inserting data for key %s: %v", key, err)
				} else {
					log.Printf("Inserted config: %s = %v\n", key, value)
				}
			}
		}

		if _, err = o.Update(&template); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func UpdateTemplateCronInfo(templateName, clusterName string, cronInfo *TemplateCronInfo) (*Template, error) {
	o := orm.NewOrm()

	// 查找指定 template
	template := &Template{}
	err := o.QueryTable("template").
		Filter("template_name", templateName).
		Filter("cluster_name", clusterName).
		One(template)
	if err != nil {
		return nil, err // 如果查找失败，返回错误
	}

	// 更新 template 字段
	template.CronSwitch = cronInfo.CronSwitch
	template.CronExpr = cronInfo.CronExpr
	template.CronFrequency = cronInfo.CronFrequency
	template.CronHour = cronInfo.CronHour
	template.CronMinute = cronInfo.CronMinute
	template.DayOfWeek = cronInfo.DayOfWeek
	template.DayOfMonth = cronInfo.DayOfMonth
	template.JobName = cronInfo.JobName
	template.IpList = cronInfo.IpList
	template.BaseIp = cronInfo.BaseIp

	// 执行更新操作
	_, err = o.Update(template)
	if err != nil {
		return nil, err // 更新失败，返回错误
	}

	return template, nil // 成功时返回更新后的 template 指针
}

func QueryScriptsByTemplateID(templateID int64) ([]string, error) {
	o := orm.NewOrm()

	template := Template{}

	// 使用指针类型加载模板
	err := o.QueryTable(new(Template)).Filter("template_id", templateID).One(&template)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, fmt.Errorf("template with ID %d not found", templateID)
		}
		return nil, fmt.Errorf("error querying template: %v", err)
	}

	// 加载与 Template 关联的所有 Script
	if _, err = o.LoadRelated(&template, "Scripts"); err != nil {
		return nil, err
	}

	// 获取 Scripts 的名称列表
	scriptNames := make([]string, 0, len(template.Scripts))
	for _, script := range template.Scripts {
		scriptNames = append(scriptNames, script.Name)
	}

	return scriptNames, nil
}

func QueryTemplateByMode(clusterName string, mode string, resource string) (interface{}, error) {
	o := orm.NewOrm()

	var templateList []Template
	_, err := o.QueryTable("template").Filter("mode", mode).
		Filter("cluster_name", clusterName).
		Filter("resource", resource). // 根据resource字段筛选
		All(&templateList)
	if err != nil {
		return nil, err
	}

	var templateNames []string
	for _, t := range templateList {
		templateNames = append(templateNames, t.TemplateName)
	}

	return templateNames, nil
}

func QueryDetailOfTemplate(templateName, clusterName string) (interface{}, error) {
	o := orm.NewOrm()

	// 查询模板
	t := &Template{TemplateName: templateName, ClusterName: clusterName}
	err := o.Read(t, "template_name", "cluster_name")
	if err != nil {
		return nil, err
	}

	// 加载与 Template 关联的所有 Script
	if _, err = o.LoadRelated(t, "Scripts"); err != nil {
		return nil, err
	}

	domainDict := make(map[string][]map[string]interface{})
	for _, script := range t.Scripts {
		scriptName := script.ChName
		domain := script.Domain

		params, err := QueryParamsByScriptName(t.TemplateID, scriptName)
		if err != nil {
			log.Printf("Error fetching params for script %s: %v\n", scriptName, err)
			continue
		}

		// 构建包含脚本名和参数的 map
		scriptDetails := map[string]interface{}{
			"scriptName": scriptName,
			"params":     params,
		}

		// 初始化 domain 的数组
		if _, exists := domainDict[domain]; !exists {
			domainDict[domain] = []map[string]interface{}{}
		}

		// 将脚本信息添加到 domain 的数组中
		domainDict[domain] = append(domainDict[domain], scriptDetails)
	}

	var cronTime string

	if t.CronMinute != "" {
		if t.CronFrequency == "minutely" {
			cronTime = fmt.Sprintf("%02s", t.CronMinute)
		} else {
			cronTime = fmt.Sprintf("%02s:%02s", t.CronHour, t.CronMinute)
		}
	}

	// 构建结果
	result := map[string]interface{}{
		"template":    t.TemplateName,
		"description": t.Description,
		"mode":        t.Mode,
		"resource":    t.Resource,
		"needCron":    t.CronSwitch,
		"frequency":   t.CronFrequency,
		"cronTime":    cronTime,
		"dayOfWeek":   t.DayOfWeek,
		"dayOfMonth":  t.DayOfMonth,
		"jobName":     t.JobName,
		"ipList":      ParseIPList(t.IpList),
		"baseIP":      t.BaseIp,
		"scriptList":  domainDict,
	}

	return result, nil
}

func ParseIPList(ipList string) []string {
	if ipList == "" {
		return []string{} // 返回空切片
	}
	return strings.Split(ipList, ",")
}
