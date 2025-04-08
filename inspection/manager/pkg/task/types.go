//go:generate deepcopy-gen --bounding-dirs ./pkg/task --output-file zz_generated/deepcopy.go

package task

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// TaskSpec 定义任务的规格
type TaskSpec struct {
	JobId        string `json:"job_id"`
	JobName      string `json:"job_name"`
	UserName     string `json:"user_name"`
	ClusterName  string `json:"cluster_name"`
	IsCron       bool   `json:"is_cron"`
	TemplateName string `json:"template_name"`
	TemplateID   int64  `json:"template_id"`
	IpList       string `json:"ip_list"`
	BaseIP       string `json:"base_ip"`
	Mode         string `json:"mode"`
	Resource     string `json:"resource"`
	Status       string `json:"status"`
	CreateTime   string `json:"create_time"`
	FinishTime   string `json:"finish_time"`
}

// TaskStatus 定义任务的状态
type TaskStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}
