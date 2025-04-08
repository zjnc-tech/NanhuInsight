package templates

type CreateRequest struct {
	TemplateName string        `json:"template"`
	Description  string        `json:"description"`
	Mode         string        `json:"mode"`
	Resource     string        `json:"resource"`
	NeedCron     bool          `json:"needCron"`
	ScriptList   []ScriptParam `json:"scriptList"`
	Frequency    string        `json:"frequency"`
	Hour         string        `json:"hour"`
	Minute       string        `json:"minute"`
	DayOfWeek    string        `json:"dayOfWeek"`
	DayOfMonth   string        `json:"dayOfMonth"`
	JobName      string        `json:"jobName"`
	IPList       []string      `json:"IPList,omitempty"`
	BaseIP       string        `json:"baseIP,omitempty"`
}

type ScriptParam struct {
	ScriptName string                 `json:"scriptName"`       // 脚本名称
	Params     map[string]interface{} `json:"params,omitempty"` // 参数键值对
}
