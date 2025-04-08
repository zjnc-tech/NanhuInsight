package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type JobInfo struct {
	ID           int64  `json:"id" orm:"column(id);pk;auto"`
	JobId        string `json:"job_id" orm:"column(job_id);"`
	JobName      string `json:"job_name" orm:"size(128)"`
	UserName     string `json:"user_name" orm:"size(128)"`
	ClusterName  string `json:"cluster_name" orm:"size(128)"`
	IsCron       bool   `json:"is_cron" orm:"default(false)"`
	TemplateName string `json:"template_name" orm:"size(128)"`
	TemplateID   int64  `json:"template_id" orm:"column(template_id)"`
	IpList       string `json:"ip_list" orm:"type(text);null"`
	BaseIp       string `json:"base_ip" orm:"size(128);null"`
	Mode         string `json:"mode" orm:"size(20);default(regular)"`
	Resource     string `json:"resource" orm:"size(128)"`
	Status       string `json:"status" orm:"size(20)"`
	CreateTime   string `json:"create_time" orm:"size(128)"`
	FinishTime   string `json:"finish_time" orm:"size(128)"`
}

func InitJobInfo() {
	orm.RegisterModel(new(JobInfo))
}

// AddJobInfo insert a new JobInfo into database
func AddJobInfo(jobInfo *JobInfo) error {
	o := orm.NewOrm()

	if _, err := o.Insert(jobInfo); err != nil {
		return err
	}
	return nil
}

// GetJobInfo gets JobInfo from database
func GetJobInfo(jobId string) (*JobInfo, error) {
	o := orm.NewOrm()

	var jobInfo JobInfo

	// Query the database
	err := o.QueryTable("job_info").Filter("job_id", jobId).One(&jobInfo)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			log.Printf("Job not found with job ID: %s", jobId)
		} else {
			log.Printf("Error fetching job info for job ID: %s, error: %s", jobId, err.Error())
		}
		return nil, err
	}

	// Successfully found the job
	return &jobInfo, nil
}

// QueryJobCountByStatus query if there is JobInfo in certain cluster and in ongoing or creating status.
func QueryJobCountByStatus(cluster string) bool {
	statusList := []string{"ongoing", "creating"}
	o := orm.NewOrm()

	// 构建查询条件
	qs := o.QueryTable("job_info")
	qs = qs.Filter("cluster_name", cluster)
	qs = qs.Filter("status__in", statusList) // 使用__in过滤器进行IN查询

	// 统计符合条件的记录数量
	num, err := qs.Count()
	if err != nil {
		log.Printf("query job error: %v", err)
		return false
	}

	return num > 0
}

// UpdateJobInfo update status and finishTime of JobInfo
func UpdateJobInfo(jobId, status, finishTime string, templateId int64) (err error) {
	o := orm.NewOrm()
	var jobInfo JobInfo

	qs := o.QueryTable("job_info").Filter("job_id", jobId)
	if err = qs.One(&jobInfo); err == nil {
		jobInfo.Status = status
		jobInfo.FinishTime = finishTime
		jobInfo.TemplateID = templateId
		if _, err = o.Update(&jobInfo); err == nil {
			log.Printf("Job status updated to '%s' successfully for job ID: %s", status, jobId)
			return nil
		} else {
			log.Printf("Failed to update job status for job ID: %s, error: %s", jobId, err.Error())
			return err
		}
	} else {
		log.Printf("Job not found with job ID: %s, error: %s", jobId, err.Error())
		return err
	}
}

func QueryJobInfoById(jobId string, field string) (interface{}, error) {
	o := orm.NewOrm()

	var jobInfo JobInfo
	err := o.QueryTable("job_info").Filter("job_id", jobId).One(&jobInfo)

	if errors.Is(orm.ErrNoRows, err) {
		return nil, fmt.Errorf("job info with id %s does not exist", jobId)
	} else if err != nil {
		return nil, err
	}

	// 根据指定的字段返回结果
	switch field {
	case "job_name":
		return jobInfo.JobName, nil
	case "create_time":
		return jobInfo.CreateTime, nil
	case "finish_time":
		return jobInfo.FinishTime, nil
	case "cluster_name":
		return jobInfo.ClusterName, nil
	default:
		return nil, fmt.Errorf("invalid field specified")
	}
}

func GetJobRecords(query map[string]string, pageID int64, pageSize int64,
	sort string) (jobRecords []JobInfo, totalNum int64, err error) {
	o := orm.NewOrm()

	qs := o.QueryTable("job_info")

	// 可按需添加查询条件
	for key, value := range query {
		if key == "createTimeRange" {
			var startTime, endTime *time.Time
			// 解码时间范围
			startTime, endTime, err = decodeTimeRange(value)
			if err != nil {
				fmt.Println("Error decoding time range:", err)
				return
			}

			// 根据解析结果动态添加过滤条件
			if startTime != nil {
				qs = qs.Filter("create_time__gte", *startTime)
			}
			if endTime != nil {
				qs = qs.Filter("create_time__lte", *endTime)
			}
		} else if key == "finishedTimeRange" {
			var startTime, endTime *time.Time
			// 解码时间范围
			startTime, endTime, err = decodeTimeRange(value)
			if err != nil {
				fmt.Println("Error decoding time range:", err)
				return
			}

			// 根据解析结果动态添加过滤条件
			if startTime != nil {
				qs = qs.Filter("finish_time__gte", *startTime)
			}
			if endTime != nil {
				qs = qs.Filter("finish_time__lte", *endTime)
			}
		} else {
			// 处理其他字段
			qs = qs.Filter(key, value)
		}
	}

	qs = qs.OrderBy(sort)

	// 统计总记录数（用于计算总页数）
	totalNum, err = qs.Count()
	if err != nil {
		return nil, 0, err
	}

	_, err = qs.Limit(pageSize, (pageID-1)*pageSize).All(&jobRecords)
	if err != nil {
		return nil, 0, err
	}

	return jobRecords, totalNum, nil
}

func decodeTimeRange(encodedRange string) (*time.Time, *time.Time, error) {
	var timeRange [2]*time.Time

	decodedRange, err := url.QueryUnescape(encodedRange)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode URL: %w", err)
	}

	var timeStrings []string
	err = json.Unmarshal([]byte(decodedRange), &timeStrings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	layout := "2006-01-02 15:04:05"
	loc, err := time.LoadLocation("Asia/Shanghai") // 确保时区与数据库一致
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load time location: %w", err)
	}

	if len(timeStrings) > 0 && timeStrings[0] != "" {
		startTime, err := time.ParseInLocation(layout, timeStrings[0], loc)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		timeRange[0] = &startTime
	}
	if len(timeStrings) > 1 && timeStrings[1] != "" {
		endTime, err := time.ParseInLocation(layout, timeStrings[1], loc)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse end time: %w", err)
		}
		timeRange[1] = &endTime
	}

	return timeRange[0], timeRange[1], nil
}
