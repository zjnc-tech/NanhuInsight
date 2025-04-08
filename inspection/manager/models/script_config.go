package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
)

type ScriptConfig struct {
	ID          int64  `json:"id" orm:"column(id);pk;auto"`
	ScriptName  string `json:"script_name" orm:"column(script_name)"`
	TemplateID  int64  `json:"template_id" orm:"column(template_id)"`
	ConfigName  string `json:"config_name" orm:"column(config_name);size(100)"`
	ConfigValue string `json:"config_value" orm:"column(config_value);size(255)"`
}

func InitScriptConfig() {
	orm.RegisterModel(new(ScriptConfig))
}

func AddScriptConfig(config *ScriptConfig) error {
	o := orm.NewOrm()

	if _, err := o.Insert(config); err != nil {
		return err
	}
	return nil
}

func QueryParamsByScriptName(templateID int64, scriptName string) (map[string]interface{}, error) {
	o := orm.NewOrm()
	var params []ScriptConfig

	// 查询参数记录
	_, err := o.QueryTable("script_config").Filter("script_name", scriptName).
		Filter("template_id", templateID).All(&params)
	if err != nil {
		return nil, fmt.Errorf("error querying params for script %s: %w", scriptName, err)
	}

	// 组装参数为 map[string]interface{}
	paramMap := make(map[string]interface{})
	for _, param := range params {
		paramMap[param.ConfigName] = param.ConfigValue
	}

	return paramMap, nil
}
