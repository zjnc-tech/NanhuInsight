package job

type CreateJobRequest struct {
	JobName      string   `json:"jobName"`
	TemplateName string   `json:"template"`
	IsCron       bool     `json:"isCron"`
	Mode         string   `json:"mode"`
	Resource     string   `json:"resource"`
	IpList       []string `json:"ipList,omitempty"`
	BaseIP       string   `json:"baseIP,omitempty"`
}
