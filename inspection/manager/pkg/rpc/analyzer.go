package rpc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/alert"
	pb "infrahi/backend/inspection-manager/pkg/proto"
)

func GetNodesFromKey(input string) []string {
	// 正则表达式匹配括号内的内容
	re := regexp.MustCompile(`\((.*?)\)`)
	matches := re.FindAllStringSubmatch(input, -1)

	var nodes []string
	for _, match := range matches {
		nodes = append(nodes, match[1])
	}
	return nodes
}

func processCaseResult(checkResult *pb.CheckResult, nodeNum int, caseName, jobID, timeCostRounded, clusterName string) {
	// 创建 TestInfo 基本结构
	createTestInfo := func(result string, healthyNum, unhealthyNum, criticalNum, unknownNum, timeoutNum,
		totalNum int) models.TestInfo {
		return models.TestInfo{
			JobId:        jobID,
			CaseName:     caseName,
			HealthyNum:   healthyNum,
			UnhealthyNum: unhealthyNum,
			CriticalNum:  criticalNum,
			UnknownNum:   unknownNum,
			TimeoutNum:   timeoutNum,
			TotalNum:     totalNum,
			TimeCost:     timeCostRounded,
			Result:       result,
		}
	}

	// 如果 checkResult 为空，直接记录失败结果
	if checkResult == nil {
		testInfo := createTestInfo("failed", 0, 0, 0, 0, 0, nodeNum)
		if err := models.AddTestInfo(&testInfo); err != nil {
			log.Printf("Failed to add test info to DB: %v", err)
		}
		return
	}

	// 记录日志
	if err := writeCaseLog(checkResult.LogsResult, jobID, caseName); err != nil {
		log.Printf("Write test result log error: %v", err)
	}

	// 初始化统计数据
	caseResult := checkResult.CaseResult
	totalNum, healthyNum, unhealthyNum, criticalNum, unknownNum, timeoutNum := 0, 0, 0, 0, 0, 0

	// 遍历 caseResult 进行统计
	for key, value := range caseResult {
		nodes := GetNodesFromKey(key)
		count := len(nodes)
		totalNum += count

		switch value {
		case 0: // Healthy
			healthyNum += count
		case 1: // Critical
			criticalNum += count
		case 2: // Unhealthy
			unhealthyNum += count
		case 3: // Unknown
			unknownNum += count
		case 4: // Timeout
			timeoutNum += count
		}

		if value > 0 && value <= 4 {
			alert.CallFault(int(value), checkResult.LogsResult[key], clusterName, caseName, nodes)
		}
	}

	// 计算结果状态
	result := "not passed"
	if criticalNum == 0 {
		result = "passed"
	}

	// 保存统计结果
	testInfo := createTestInfo(result, healthyNum, unhealthyNum, criticalNum, unknownNum, timeoutNum, totalNum)
	if err := models.AddTestInfo(&testInfo); err != nil {
		log.Printf("Failed to add test info to DB: %v", err)
	}
}

func ensureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	return nil
}

func writeCaseLog(logsResult map[string]string, jobID, caseName string) error {
	logPath, err := beego.AppConfig.String("log_path")
	directory := filepath.Join(logPath, jobID)

	// 确保目录存在
	if err = ensureDirectoryExists(directory); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", directory, err)
	}

	// 设置文件路径
	fileBaseName := strings.ReplaceAll(caseName, " ", "_")
	logFile := filepath.Join(directory, fileBaseName+".log")

	// 打开文件，如果文件不存在则创建，如果存在则覆盖
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file %s: %v", logFile, err)
	}
	defer file.Close()

	// 将日志写入文件
	for key, value := range logsResult {
		_, err = file.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		if err != nil {
			return fmt.Errorf("failed to write log to file %s: %v", logFile, err)
		}
	}

	return nil
}
