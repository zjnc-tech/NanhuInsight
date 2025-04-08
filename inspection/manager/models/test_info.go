package models

import (
	"fmt"
	"log"

	"github.com/beego/beego/v2/client/orm"
)

type TestInfo struct {
	ID           int64  `json:"id" orm:"column(id);pk;auto"`
	JobId        string `json:"job_id" orm:"column(job_id);size(128));"`
	CaseName     string `json:"case_name" orm:"size(128)"`
	HealthyNum   int    `json:"healthy_num" orm:"size(128)"`
	UnhealthyNum int    `json:"unhealthy_num" orm:"size(128)"`
	CriticalNum  int    `json:"critical_num"`
	UnknownNum   int    `json:"unknown_num"`
	TimeoutNum   int    `json:"timeout_num"`
	TotalNum     int    `json:"total_num"`
	TimeCost     string `json:"time_cost"`
	Result       string `json:"result" orm:"size(128)"`
}

func InitTestInfo() {
	orm.RegisterModel(new(TestInfo))
}

// AddTestInfo insert a new TestInfo into database
func AddTestInfo(testInfo *TestInfo) error {
	o := orm.NewOrm()

	if _, err := o.Insert(testInfo); err != nil {
		return err
	}
	return nil
}

func GetTestRecords(jobId string) ([]TestInfo, error) {
	o := orm.NewOrm()

	// 构建查询条件
	qs := o.QueryTable("test_info")
	qs = qs.Filter("JobId", jobId)

	// 创建 TestInfo 切片
	var testInfos []TestInfo

	// 执行查询并将结果存入 testInfos
	_, err := qs.All(&testInfos)
	if err != nil {
		return nil, err
	}
	if len(testInfos) == 0 {
		return nil, fmt.Errorf("job %s not exist", jobId)
	}

	return testInfos, nil
}

func GetTestRecordByCase(jobId, caseName string) (*TestInfo, error) {
	o := orm.NewOrm()

	var testInfo TestInfo

	err := o.QueryTable("test_info").
		Filter("JobId", jobId).
		Filter("CaseName", caseName).
		One(&testInfo)

	if err != nil {
		return nil, err
	}

	// 返回查询到的记录和 nil 错误
	return &testInfo, nil
}

func QueryCountByTestResult(jobId, result string) int {
	o := orm.NewOrm()

	// 构建查询条件
	qs := o.QueryTable("test_info")
	qs = qs.Filter("JobId", jobId)
	qs = qs.Filter("Result", result)

	// 统计符合条件的记录数量
	num, err := qs.Count()
	if err != nil {
		log.Printf("query count error: %v", err)
		return 0
	}

	return int(num)
}
