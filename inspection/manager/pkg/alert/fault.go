package alert

import (
	"log"
	"strconv"
	"strings"
	"time"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/curl"
)

func generateAlertData(level int, result string, clusterName string, scriptName string,
	nodes []string) map[string]interface{} {
	var alerts []map[string]interface{}

	caseID, _ := models.QueryScriptByName(scriptName, "case_id")

	for _, nodeName := range nodes {
		alert := map[string]interface{}{
			"labels": map[string]string{
				"key":        string(GenerateNodeMirrorName(nodeName, clusterName)),
				"source":     "inspectionSys",
				"businessId": strconv.Itoa(caseID.(int)),
				"btype":      ResultMap[level],
			},
			"annotations": map[string]string{
				"description": result,
			},
			"startsAt": time.Now().Format(time.RFC3339),
		}
		alerts = append(alerts, alert)
	}

	return map[string]interface{}{
		"alerts": alerts,
	}
}

func CallFault(level int, result string, clusterName string, scriptName string, nodes []string) {
	if strings.Contains(result, "allocated") {
		log.Printf("ignore allocated record with message: %s", result)
		return
	}
	data := generateAlertData(level, result, clusterName, scriptName, nodes)
	if err := curl.CallCurlPost("fault_url", data); err != nil {
		log.Printf("call fault error: %v", err)
	}
}
