package cronjob

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/job"
	"infrahi/backend/inspection-manager/pkg/templates"
)

var weekDayMap = map[string]int{
	"mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6, "sun": 7,
}

var mu sync.Mutex

var cronMap = make(map[int64]*cron.Cron)

func SyncCron() {
	templateList, err := models.GetCronTemplates()
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, templateInfo := range templateList {
		user := templateInfo.CreateUser
		if templateInfo.ModifyUser != "" {
			user = templateInfo.ModifyUser
		}
		go AddScheduleJob(&templateInfo, user, templateInfo.ClusterName)
	}
	return
}

func CreateCronJob(req templates.CreateRequest, userName string, clusterName string) error {
	cronExpr, err := constructCronExpr(req)
	if err != nil {
		return err
	}

	cronInfo := &models.TemplateCronInfo{
		CronSwitch:    req.NeedCron,
		CronFrequency: req.Frequency,
		CronExpr:      cronExpr,
		CronHour:      req.Hour,
		CronMinute:    req.Minute,
		DayOfWeek:     req.DayOfWeek,
		DayOfMonth:    req.DayOfMonth,
		JobName:       req.JobName,
		IpList:        strings.Join(req.IPList, ","),
		BaseIp:        req.BaseIP,
	}

	templateInfo, err := models.UpdateTemplateCronInfo(req.TemplateName, clusterName, cronInfo)
	if err != nil {
		return err
	}

	go AddScheduleJob(templateInfo, userName, clusterName)

	return nil
}

func DeleteCronJob(templateName, clusterName string) error {
	_ = RemoveScheduleJob(templateName, clusterName)

	cronInfo := &models.TemplateCronInfo{
		CronSwitch:    false,
		CronFrequency: "",
		CronExpr:      "",
		CronHour:      "",
		CronMinute:    "",
		DayOfWeek:     "",
		DayOfMonth:    "",
		JobName:       "",
		IpList:        "",
		BaseIp:        "",
	}

	_, err := models.UpdateTemplateCronInfo(templateName, clusterName, cronInfo)
	if err != nil {
		return err
	}

	return nil
}

func AddScheduleJob(template *models.Template, userName string, clusterName string) {
	jobParam := job.CreateJobRequest{
		JobName:      template.JobName,
		TemplateName: template.TemplateName,
		IsCron:       true,
		Mode:         template.Mode,
		Resource:     template.Resource,
		IpList:       strings.Split(template.IpList, ","),
		BaseIP:       template.BaseIp,
	}

	c := cron.New()

	// 添加定时任务
	_, err := c.AddFunc(template.CronExpr, func() {
		fmt.Printf("Executing job at %s\n", time.Now().Format(time.RFC3339))
		job.CallCreateUrl(jobParam, userName, clusterName)
	})
	if err != nil {
		log.Printf("Add cron job err:%v", err)
	}

	mu.Lock()
	cronMap[template.TemplateID] = c
	mu.Unlock()

	c.Start()
	defer c.Stop()

	// 保持主程序运行
	select {}
}

func RemoveScheduleJob(templateName string, clusterName string) error {
	templateId, _ := models.GetTemplateIdByName(templateName, clusterName)

	c, exists := cronMap[templateId]
	if !exists {
		return fmt.Errorf("cron job of template %s in cluster %s does not exist", templateName, clusterName)
	}
	c.Stop()

	delete(cronMap, templateId)

	return nil
}

func constructCronExpr(req templates.CreateRequest) (string, error) {
	if req.Mode != "regular" {
		return "", fmt.Errorf("only regular templates can be used for creating scheduled tasks")
	}
	if req.Frequency == "" {
		return "", fmt.Errorf("frequency is required")
	}

	minute, err := strconv.Atoi(req.Minute)
	if err != nil || minute < 0 || minute > 59 {
		return "", fmt.Errorf("invalid minute, must be an integer between 0 and 59")
	}

	var hour int
	if req.Frequency == "daily" || req.Frequency == "weekly" || req.Frequency == "monthly" {
		hour, err = strconv.Atoi(req.Hour)
		if err != nil || hour < 0 || hour > 23 {
			return "", fmt.Errorf("invalid hour, must be an integer between 0 and 23")
		}
	}

	var cronExpr string
	switch req.Frequency {
	case "minutely":
		now := time.Now()
		minuteNow := now.Minute()

		cronExpr = fmt.Sprintf("%d/%d * * * *", minuteNow+1, minute)

	case "hourly":
		cronExpr = fmt.Sprintf("%d * * * *", minute)

	case "daily":
		cronExpr = fmt.Sprintf("%d %d * * *", minute, hour)

	case "weekly":
		if req.DayOfWeek == "" {
			return "", fmt.Errorf("missing dayOfWeek for weekly frequency")
		}
		dayOfWeek, exists := weekDayMap[req.DayOfWeek]
		if !exists {
			return "", fmt.Errorf("invalid day_of_week, must be one of: 'mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'")
		}
		cronExpr = fmt.Sprintf("%d %d * * %d", minute, hour, dayOfWeek)

	case "monthly":
		if req.DayOfMonth == "" {
			return "", fmt.Errorf("missing dayOfMonth for monthly frequency")
		}
		dayOfMonth, err := strconv.Atoi(req.DayOfMonth)
		if err != nil || dayOfMonth < 1 || dayOfMonth > 31 {
			return "", fmt.Errorf("invalid dayOfMonth, must be an integer between 1 and 31")
		}
		cronExpr = fmt.Sprintf("%d %d %d * *", minute, hour, dayOfMonth)

	default:
		return "", fmt.Errorf("unsupported frequency: %s", req.Frequency)
	}

	return cronExpr, nil
}
