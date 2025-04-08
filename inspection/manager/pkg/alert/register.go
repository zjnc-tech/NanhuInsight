package alert

import (
	"log"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/curl"
)

var ResultMap = map[int]string{
	1: "故障",
	2: "不健康",
	3: "未知",
	4: "超时",
}

var LevelMap = map[string]string{
	"故障":  "critical",
	"不健康": "serious",
	"未知":  "unknown",
	"超时":  "unknown",
}

func generateRegisterData(clusterName string, scriptNames []string) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, scriptName := range scriptNames {
		for key, value := range LevelMap {
			chName, _ := models.QueryScriptByName(scriptName, "ch_name")
			caseID, _ := models.QueryScriptByName(scriptName, "case_id")
			domain, _ := models.QueryScriptByName(scriptName, "domain")
			description, _ := models.QueryScriptByName(scriptName, "detail")
			caseData := map[string]interface{}{
				"businessId": caseID,
				"btype":      key,
				"cluster":    clusterName,
				"desc":       description,
				"name":       chName,
				"type":       "alert",
				"level":      value,
				"region":     domain,
				"rtype":      "node",
				"source":     "inspectionSys",
			}
			data = append(data, caseData)
		}
	}
	return data
}

func RegisterAlertRule(clusterName string, scriptNames []string) {
	data := generateRegisterData(clusterName, scriptNames)
	if err := curl.CallCurlPost("register_url", data); err != nil {
		log.Printf("register alert rule error: %v", err)
	}
}
