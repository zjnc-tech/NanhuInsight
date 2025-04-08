package models

import (
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
)

type Script struct {
	CaseID       int          `json:"case_id" orm:"column(script_id);auto"`
	Name         string       `json:"name" orm:"column(name);size(100)"`
	ChName       string       `json:"ch_name" orm:"column(ch_name);size(100)"`
	Cluster      string       `json:"cluster" orm:"type(text);null"`
	Detail       string       `json:"detail" orm:"column(detail);size(255)"`
	Domain       string       `json:"domain" orm:"column(domain);size(50)"`
	Mode         string       `json:"mode" orm:"column(mode);size(50)"`
	Category     string       `json:"category" orm:"column(category);size(50)"`
	Templates    []*Template  `orm:"reverse(many)"`
	ScriptParams []*ParamInfo `json:"params" orm:"reverse(many)"`
}

type ParamInfo struct {
	Id           int     `orm:"auto"`
	CaseID       *Script `json:"case_id" orm:"rel(fk);column(case_id)"`
	Required     bool    `json:"required" orm:"size(50)"`
	ParamType    string  `json:"param_type" orm:"size(50)"`
	ParamName    string  `json:"param_name" orm:"size(100)"`
	DefaultValue string  `json:"default_value" orm:"type(text)"`
}

type ParamResponse struct {
	Required     bool   `json:"isRequired"`
	ParamType    string `json:"paramType"`
	ParamName    string `json:"paramName"`
	DefaultValue string `json:"defaultValue"`
}

func InitScript() {
	orm.RegisterModel(new(Script), new(ParamInfo))
}

func AddScript(script *Script) error {
	o := orm.NewOrm()

	// 尝试查找是否已有记录
	existing := Script{CaseID: script.CaseID}
	if err := o.Read(&existing); err == nil {
		// 如果找到记录，先删除
		if _, err := o.Delete(&existing); err != nil {
			return fmt.Errorf("failed to delete existing script: %v", err)
		}
	} else if !errors.Is(err, orm.ErrNoRows) {
		// 如果查询出现其他错误（不是没有找到记录）
		return fmt.Errorf("failed to check existing script: %v", err)
	}

	// 插入新记录
	_, err := o.Insert(script)
	if err != nil {
		return fmt.Errorf("failed to insert new script: %v", err)
	}

	// 删除旧的 Params 并插入新的 Params
	if _, err := o.QueryTable(new(ParamInfo)).Filter("case_id", script.CaseID).Delete(); err != nil {
		return fmt.Errorf("failed to delete existing params: %v", err)
	}

	if len(script.ScriptParams) > 0 {
		for _, param := range script.ScriptParams {
			param.CaseID = script // 关联外键
			_, err := o.Insert(param)
			if err != nil {
				return fmt.Errorf("failed to insert param: %v", err)
			}
		}
	}

	return nil
}

func DeleteScript(scriptName string) error {
	o := orm.NewOrm()

	// 查找并删除对应的 Script 记录
	_, err := o.QueryTable(new(Script)).Filter("name", scriptName).Delete()
	if err != nil {
		return fmt.Errorf("error deleting script from database: %v", err)
	}

	_, err = o.QueryTable(new(ParamInfo)).Filter("case_id__name", scriptName).Delete()
	if err != nil {
		return fmt.Errorf("error deleting param of %s from database: %v", scriptName, err)
	}

	return nil
}

func QueryScriptByName(name string, field string) (interface{}, error) {
	o := orm.NewOrm()

	var script Script
	err := o.QueryTable("script").Filter("name", name).One(&script)

	if errors.Is(orm.ErrNoRows, err) {
		return nil, fmt.Errorf("test case with name %s does not exist", name)
	} else if err != nil {
		return nil, err
	}

	// 根据指定的字段返回结果
	switch field {
	case "category":
		return script.Category, nil
	case "domain":
		return script.Domain, nil
	case "detail":
		return script.Detail, nil
	case "ch_name":
		return script.ChName, nil
	case "case_id":
		return script.CaseID, nil
	default:
		return nil, fmt.Errorf("invalid field specified")
	}
}

func QueryScriptNameByChName(chName string) (string, error) {
	o := orm.NewOrm()

	var script Script
	err := o.QueryTable("script").Filter("ch_name", chName).One(&script)

	if errors.Is(orm.ErrNoRows, err) {
		return "", fmt.Errorf("script with ch name %s does not exist", chName)
	} else if err != nil {
		return "", err
	}

	return script.Name, nil
}

func AssociateScript(template *Template, chName string) error {
	o := orm.NewOrm()

	var script Script
	err := o.QueryTable("script").Filter("ch_name", chName).One(&script)
	if errors.Is(orm.ErrNoRows, err) {
		return fmt.Errorf("script with ch name %s does not exist", chName)
	} else if err != nil {
		return err
	}

	_, err = o.QueryM2M(template, "Scripts").Add(script)
	if err != nil {
		return fmt.Errorf("error associating test case with template: %w", err)
	}

	return nil
}

// 根据mode查询测试用例
func queryCaseNameByMode(mode, clusterName, resource string) (interface{}, error) {
	o := orm.NewOrm()
	var scripts []Script

	// 根据mode条件查询
	switch mode {
	case "regular":
		var tempScripts []Script

		// 先查询 cluster 为空的记录
		_, err := o.QueryTable(new(Script)).
			Filter("mode", mode).  // 筛选 mode 为 "regular"
			Filter("cluster", ""). // 筛选 cluster 为空
			All(&tempScripts)
		if err != nil {
			return nil, err
		}

		// 查询 cluster 包含 clusterName 的记录
		_, err = o.QueryTable(new(Script)).
			Filter("mode", mode).
			Filter("cluster__icontains", clusterName+":"+resource). // `cluster` 字段包含 clusterName:resource
			All(&scripts)
		if err != nil {
			return nil, err
		}

		// 将两个查询结果合并
		scripts = append(scripts, tempScripts...)
		return scripts, nil
	case "deep":
		var tempScripts []Script

		// 先查询 cluster 为空的记录
		_, err := o.QueryTable(new(Script)).
			Filter("cluster", ""). // 筛选 cluster 为空
			All(&tempScripts)
		if err != nil {
			return nil, err
		}

		// 查询 cluster 包含 clusterName 的记录
		_, err = o.QueryTable(new(Script)).
			Filter("cluster__icontains", clusterName+":"+resource). // `cluster` 字段包含 clusterName:resource
			All(&scripts)
		if err != nil {
			return nil, err
		}

		// 将两个查询结果合并
		scripts = append(scripts, tempScripts...)
		return scripts, nil
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}

// GetScriptsByMode 根据mode获取并分类测试用例
func GetScriptsByMode(mode, clusterName, resource string) (interface{}, error) {
	result, err := queryCaseNameByMode(mode, clusterName, resource)
	if err != nil {
		return nil, err
	}

	var domainDict = map[string][]string{
		"compute": {},
		"network": {},
		"other":   {},
		"storage": {},
	}

	cases, ok := result.([]Script)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	for _, caseItem := range cases {
		domain := caseItem.Domain
		name := caseItem.ChName

		if _, exists := domainDict[domain]; !exists {
			domain = "other"
		}

		domainDict[domain] = append(domainDict[domain], name)
	}

	return domainDict, nil
}

func GetScriptConfigs(scriptNames []string) (map[string][]ParamResponse, error) {
	result := make(map[string][]ParamResponse)

	for _, scriptName := range scriptNames {
		params, err := GetScriptConfig(scriptName)
		if err != nil {
			return nil, fmt.Errorf("error fetching config for script '%s': %v", scriptName, err)
		}
		result[scriptName] = params
	}

	return result, nil
}

func GetScriptConfig(scriptName string) ([]ParamResponse, error) {
	o := orm.NewOrm()
	var script Script
	var params []ParamInfo

	// 根据 ChName 查找记录
	err := o.QueryTable(new(Script)).
		Filter("ch_name", scriptName).
		One(&script)
	if err != nil {
		return nil, fmt.Errorf("failed to find script with ChName=%s: %v", scriptName, err)
	}

	if errors.Is(err, orm.ErrNoRows) {
		return nil, nil // 未找到
	}
	if err != nil {
		return nil, err
	}

	// 获取对应的 Params
	_, err = o.QueryTable(new(ParamInfo)).Filter("case_id", script.CaseID).All(&params)
	if err != nil {
		return nil, err
	}

	// 构造返回的数据，不包含 case_id 部分
	var response []ParamResponse
	for _, param := range params {
		response = append(response, ParamResponse{
			Required:     param.Required,
			ParamType:    param.ParamType,
			ParamName:    param.ParamName,
			DefaultValue: param.DefaultValue,
		})
	}

	return response, nil
}
