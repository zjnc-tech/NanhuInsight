package job

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/utils"
)

func CreateTaskInstance(job models.JobInfo, namespace string) error {
	client := utils.GetK8sClient()

	taskCrd := map[string]interface{}{
		"apiVersion": "isg.zjlab.io/v1",
		"kind":       "Task",
		"metadata": map[string]interface{}{
			"name":      job.JobId,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"job_id":        job.JobId,
			"job_name":      job.JobName,
			"user_name":     job.UserName,
			"cluster_name":  job.ClusterName,
			"is_cron":       job.IsCron,
			"template_name": job.TemplateName,
			"template_id":   job.TemplateID,
			"ip_list":       job.IpList,
			"base_ip":       job.BaseIp,
			"mode":          job.Mode,
			"resource":      job.Resource,
			"status":        job.Status,
			"create_time":   job.CreateTime,
			"finish_time":   job.FinishTime,
		},
	}

	taskUnstructured := &unstructured.Unstructured{Object: taskCrd}
	gvr := schema.GroupVersionResource{
		Group:    "isg.zjlab.io",
		Version:  "v1",
		Resource: "tasks",
	}

	_, err := client.Resource(gvr).Namespace(namespace).Create(context.TODO(), taskUnstructured, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Task instance in namespace %s: %w", namespace, err)
	}

	log.Printf("Task CRD instance for job %s created successfully in namespace %s!", job.JobId, namespace)
	return nil
}
