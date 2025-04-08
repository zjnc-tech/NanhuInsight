package job

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/alert"
	"infrahi/backend/inspection-manager/pkg/cluster"
	"infrahi/backend/inspection-manager/pkg/rpc"
	"infrahi/backend/inspection-manager/pkg/utils"
)

func ProcessJob(jobInfo models.JobInfo) {
	// 获取待检查节点信息
	nodesInfo, getErr := cluster.GetProcessNodeInfo(jobInfo.ClusterName, jobInfo.Mode, jobInfo.IpList,
		jobInfo.BaseIp, jobInfo.Resource)
	if getErr != nil {
		log.Printf("Nodes info get err: %v", getErr)
		if err := updateJobAndTemplate(jobInfo.JobId, jobInfo.TemplateID, "creation failed"); err != nil {
			log.Printf("updateJobAndTemplate error: %v", err)
		}
		return
	}

	log.Printf("Nodes info: %v", nodesInfo)

	if err := updateJobAndTemplate(jobInfo.JobId, jobInfo.TemplateID, "ongoing"); err != nil {
		log.Printf("process job error: %v", err)
	}

	status := executeJob(jobInfo.ClusterName, jobInfo.JobId, jobInfo.TemplateID, nodesInfo)

	if err := updateJobAndTemplate(jobInfo.JobId, jobInfo.TemplateID, status); err != nil {
		log.Printf("process job error: %v", err)
	}
	return
}

func executeJob(clusterName string, jobId string, templateID int64, nodesInfo rpc.JobNodesInfo) string {
	scriptNames, err := models.QueryScriptsByTemplateID(templateID)
	if err != nil {
		return "creation failed"
	}

	if len(scriptNames) == 0 {
		log.Printf("No script found for template id %d", templateID)
		return "creation failed"
	}

	agentAddr, err := utils.GetAgentAddress(clusterName)
	if err != nil {
		return "creation failed"
	}

	// 注册告警规则ID
	alert.RegisterAlertRule(clusterName, scriptNames)

	success := rpc.ExecuteScripts(agentAddr, clusterName, jobId, templateID, scriptNames, nodesInfo)

	if success {
		return "completed"
	}
	return "task failed"
}

func updateJobAndTemplate(jobId string, templateId int64, status string) (err error) {
	var finishTime string
	if status != "ongoing" {
		finishTime = time.Now().Format("2006-01-02 15:04:05")
	}

	if err = models.UpdateJobInfo(jobId, status, finishTime, templateId); err != nil {
		return err
	}

	UpdateTask(jobId, finishTime, status)

	if status == "ongoing" {
		return nil
	}

	if err = models.UpdateTemplateJobInfo(templateId, jobId, finishTime); err != nil {
		return err
	}

	return nil
}

func UpdateTask(jobID string, finishTime string, status string) {
	client := utils.GetK8sClient()

	taskGVR := schema.GroupVersionResource{
		Group:    "isg.zjlab.io",
		Version:  "v1",
		Resource: "tasks",
	}

	taskUnstructured, err := client.Resource(taskGVR).
		Namespace("isg").
		Get(context.TODO(), jobID, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting Task: %v", err)
	}

	taskMap := taskUnstructured.UnstructuredContent()

	spec, found := taskMap["spec"].(map[string]interface{})
	if !found {
		log.Printf("Spec field not found in Task")
		return
	}
	spec["finish_time"] = finishTime
	spec["status"] = status

	updatedTask := &unstructured.Unstructured{Object: taskMap}

	_, err = client.Resource(taskGVR).
		Namespace("isg").
		Update(context.TODO(), updatedTask, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Failed to update Task status to '%s': %v", status, err)
	} else {
		log.Printf("Task %s marked as '%s' with finish time: %s", jobID, status, spec["finish_time"])
	}

	return
}
