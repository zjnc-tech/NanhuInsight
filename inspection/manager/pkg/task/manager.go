package task

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/job"
	"infrahi/backend/inspection-manager/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func createTaskCRD() error {
	tempPath, _ := beego.AppConfig.String(beego.BConfig.RunMode + "::kube_path")
	filePath := filepath.Join(tempPath, "task_crd.yaml")

	log.Printf("from yaml Create task CRD: %v", filePath)

	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file %s: %w", filePath, err)
	}

	reader := bytes.NewReader(yamlFile)

	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)
	var crd unstructured.Unstructured
	if err := decoder.Decode(&crd); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	dynamicClient := utils.GetK8sClient()

	crdClient := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	})

	crdName := crd.GetName()

	_, err = crdClient.Get(context.TODO(), crdName, metav1.GetOptions{})
	if err == nil {
		log.Printf("CRD %s already exists, skipping creation.", crdName)
		return nil
	}

	// 如果查询出错且不是 "Not Found" 错误，则返回错误
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if CRD exists: %w", err)
	}

	_, err = crdClient.Create(context.TODO(), &crd, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create CRD: %w", err)
	}

	log.Printf("CRD task created successfully!")
	return nil
}

func Handle(task *Task) {
	if task.Spec.Status != "creating" {
		return
	}
	log.Printf("Handling task: %s", task.Spec.JobId)
	// 准备作业信息
	jobInfo := models.JobInfo{
		UserName:     task.Spec.UserName,
		ClusterName:  task.Spec.ClusterName,
		TemplateName: task.Spec.TemplateName,
		TemplateID:   task.Spec.TemplateID,
		JobName:      task.Spec.JobName,
		JobId:        task.Spec.JobId,
		Status:       "creating",
		IsCron:       task.Spec.IsCron,
		Mode:         task.Spec.Mode,
		Resource:     task.Spec.Resource,
		IpList:       task.Spec.IpList,
		BaseIp:       task.Spec.BaseIP,
		CreateTime:   task.Spec.CreateTime,
	}
	job.JobManager.EnqueueJob(jobInfo)
	return
}

func WatchJobTasks() {
	if beego.BConfig.RunMode == "dev" {
		return
	}

	log.Printf("Start watching tasks")

	for {
		client := utils.GetK8sClient()

		taskGVR := schema.GroupVersionResource{
			Group:    "isg.zjlab.io",
			Version:  "v1",
			Resource: "tasks",
		}

		// 创建 informer 来监听 Task CRD
		informer, _ := client.Resource(taskGVR).Namespace("isg").Watch(context.TODO(), metav1.ListOptions{})
		for event := range informer.ResultChan() {
			switch event.Type {
			case "ADDED":
				// 处理新增的 Task
				taskUnstructured, ok := event.Object.(*unstructured.Unstructured)
				if !ok {
					log.Printf("Error: Failed to cast event object to *unstructured.Unstructured")
					continue
				}

				// 转换 unstructured 对象为 Task 类型
				var task Task
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(taskUnstructured.Object, &task)
				if err != nil {
					log.Printf("Error: Failed to convert unstructured object to Task type: %v", err)
					continue
				}
				Handle(&task)

			case "DELETED":
				// 处理删除的 Task
				log.Printf("Task Deleted: %s", event.Object)
			}

			if event.Object == nil {
				log.Println("Watcher channel closed, restarting watch...")
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func init() {
	if beego.BConfig.RunMode == "dev" {
		fmt.Println("CRD not supported in dev mode yet")
		return
	}

	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		log.Printf("unable to register Task scheme: %v\n", err)
	}
	log.Printf("AddToScheme successfully")

	if err := createTaskCRD(); err != nil {
		log.Printf("Failed to create task CRD: %v", err)
	}
}
