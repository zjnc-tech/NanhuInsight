package job

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/models"
)

var JobManager = NewJobQueueManager()

type QueueManager struct {
	queues    map[string]chan models.JobInfo
	processes map[string]bool
	mu        sync.Mutex
}

// NewJobQueueManager 初始化 QueueManager
func NewJobQueueManager() *QueueManager {
	return &QueueManager{
		queues:    make(map[string]chan models.JobInfo),
		processes: make(map[string]bool), // 记录 processor 是否在运行
	}
}

// EnqueueJob 向队列添加任务
func (m *QueueManager) EnqueueJob(job models.JobInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果队列不存在，则创建
	if _, exists := m.queues[job.ClusterName]; !exists {
		m.queues[job.ClusterName] = make(chan models.JobInfo, 100)
	}

	// 如果当前 cluster 没有 processor 运行，启动一个新的
	if !m.processes[job.ClusterName] {
		m.processes[job.ClusterName] = true
		go m.startJobProcessor(job.ClusterName)
	}

	// 发送任务
	m.queues[job.ClusterName] <- job
	log.Printf("enqueued job %s for cluster %s", job.JobId, job.ClusterName)
}

// 处理队列中的任务（每个 cluster 一个独立的 processor）
func (m *QueueManager) startJobProcessor(clusterName string) {
	log.Printf("Starting job processor in cluster: %s", clusterName)

	for {
		select {
		case job, ok := <-m.queues[clusterName]:
			if !ok {
				// 队列被关闭，退出 Goroutine
				log.Printf("Job queue for cluster %s is closed. Stopping processor.", clusterName)
				m.mu.Lock()
				delete(m.processes, clusterName)
				m.mu.Unlock()
				return
			}
			ProcessJob(job) // 串行执行任务
		case <-time.After(10 * time.Minute): // 防止 Goroutine 被空闲阻塞
			log.Printf("No jobs for cluster %s, processor sleeping.", clusterName)
		}
	}
}

func CallCreateUrl(jobParam CreateJobRequest, userName string, clusterName string) {
	// 创建请求头
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"x-username":   {userName},
		"x-cluster":    {clusterName},
	}

	// 将 jobParam 转换为 JSON
	jsonData, err := json.Marshal(jobParam)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return
	}

	// 创建 POST 请求
	req, err := http.NewRequest("POST", "http://localhost:10086/inspection/api/v1/create_job",
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	// 设置请求头
	req.Header = headers

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	// 记录响应内容
	log.Printf("create job response content: %s", respBody.String())
}

func CreateJob(jobInfo models.JobInfo) error {
	if beego.BConfig.RunMode == "dev" {
		fmt.Println("CRD not supported in dev mode")
		JobManager.EnqueueJob(jobInfo)
		return nil
	}

	// 调用 Kubernetes API 创建 CRD
	return CreateTaskInstance(jobInfo, "isg")
}
