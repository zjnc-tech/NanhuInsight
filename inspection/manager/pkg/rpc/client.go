package rpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"infrahi/backend/inspection-manager/models"
	pb "infrahi/backend/inspection-manager/pkg/proto"
	"infrahi/backend/inspection-manager/pkg/scripts"
)

var fileMutex sync.Mutex

func transferAndExecute(stub pb.ScriptTransferClient, caseName string, jobID string, templateID int64,
	clusterName string, nodesInfo JobNodesInfo) (err error, result *pb.CheckResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	fileMutex.Lock()
	defer fileMutex.Unlock()

	// 获取文件名称
	fileName := filepath.Base(caseName)

	// 读取脚本文件
	fileData, err := os.ReadFile(caseName)
	if err != nil {
		log.Printf("failed to read script file: %v", err)
		return err, nil
	}

	// 调用 TransferScript 方法
	transferResponse, err := stub.TransferScript(ctx, &pb.TransferRequest{
		FileName: fileName,
		FileData: fileData,
	})

	if err != nil {
		log.Printf("failed to transfer script: %v", err)
		return err, nil
	}

	if !transferResponse.Success {
		log.Printf("transfer script failed, reason: %v", transferResponse.Message)
		return fmt.Errorf("transfer script failed, reason: %v", transferResponse.Message), nil
	}

	params, err := getScriptParams(fileName, templateID)
	if err != nil {
		log.Printf("failed to get script params: %v", err)
		return err, nil
	}

	// 调用 Execute 方法
	executeResponse, err := stub.Execute(ctx, &pb.ExecuteRequest{
		JobId:        jobID,
		ProcessNodes: nodesInfo.ProcessNodes,
		BaseNode:     nodesInfo.BaseNode,
		ScriptName:   caseName,
		Params:       params,
		ClusterName:  clusterName,
	})

	if err != nil {
		log.Printf("failed to execute script: %v", err)
		return err, nil
	}

	if executeResponse.Success {
		return nil, executeResponse.Output
	} else {
		log.Printf("%s execute failed, reason: %v", fileName, executeResponse.Message)
		return fmt.Errorf("%s execute failed, reason: %v", fileName, executeResponse.Message), nil
	}
}

func executeCaseByRPC(stub pb.ScriptTransferClient, clusterName string, caseName string, jobID string,
	templateID int64, nodesInfo JobNodesInfo) bool {
	ret := true
	// 压缩文件夹
	zipName, zipErr := scripts.ZipScriptFolder(caseName)
	if zipErr != nil {
		log.Printf("error: %v", zipErr)
		return false
	}

	// 记录执行开始时间
	startTime := time.Now()
	log.Printf("execute job %s script %s started at %s", jobID, caseName, startTime.Format("2006-01-02 15:04:05"))

	err, result := transferAndExecute(stub, zipName, jobID, templateID, clusterName, nodesInfo)
	if err != nil {
		ret = false
		log.Printf("execute case %s failed: %v", caseName, err)
	}

	// 记录执行结束时间
	endTime := time.Now()
	timeCost := endTime.Sub(startTime).Seconds()
	timeCostRounded := fmt.Sprintf("%.2f", timeCost)

	log.Printf("execute job %s script %s ended at %s", jobID, caseName, endTime.Format("2006-01-02 15:04:05"))

	// 处理结果
	processCaseResult(result, len(nodesInfo.ProcessNodes), caseName, jobID, timeCostRounded, clusterName)

	log.Printf("execute job %s script %s processed at %s", jobID, caseName, endTime.Format("2006-01-02 15:04:05"))

	return ret
}

func ExecuteScripts(agentAddr string, clusterName string, jobID string, templateID int64, caseNames []string,
	nodesInfo JobNodesInfo) bool {
	ret := true

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建连接
	conn, err := grpc.DialContext(ctx, agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建 gRPC 客户端
	client := pb.NewScriptTransferClient(conn)

	// 执行每个用例
	for _, caseName := range caseNames {
		result := executeCaseByRPC(client, clusterName, caseName, jobID, templateID, nodesInfo)
		if result != true {
			ret = false
		}
	}

	return ret
}

func getScriptParams(fileName string, templateID int64) (params map[string]string, err error) {
	scriptName := strings.TrimSuffix(fileName, ".zip")
	chName, err := models.QueryScriptByName(scriptName, "ch_name")
	if err != nil {
		log.Printf("failed to query script by name %s: %v", scriptName, err)
	}
	scriptChName := chName.(string)

	scriptParams, err := models.QueryParamsByScriptName(templateID, scriptChName)
	if err != nil {
		log.Printf("failed to query params for scirpt %s: %v", scriptChName, err)
	}

	result := make(map[string]string)
	for key, value := range scriptParams {
		result[key] = fmt.Sprintf("%v", value) // 将值格式化为字符串
	}
	return result, nil
}

func GetResourceFromAgent(agentAddr string, IPList []string) (map[string][]string, error) {
	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建连接
	conn, err := grpc.DialContext(ctx, agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to agent: %v", err)
		return nil, err
	}
	defer conn.Close()

	// 创建 gRPC 客户端
	stub := pb.NewScriptTransferClient(conn)

	// 调用 GetResource 方法
	response, err := stub.GetResource(ctx, &pb.GetResourceRequest{
		IpList: IPList,
	})

	if err != nil {
		log.Printf("Failed to get resource: %v", err)
		return nil, err
	}

	if !response.Success {
		err := fmt.Errorf("get resource failed, reason: %v", response.Message)
		log.Printf("Error resource response: %v", err)
		return nil, err
	}

	// 初始化 nodeMap
	nodeMap := make(map[string][]string)

	// 遍历 response.Output.Node，将 CardType 转换为字符串
	for nodeName, cardType := range response.Output.Node {
		cardTypeStr := CardTypeToString[cardType]
		nodeMap[cardTypeStr] = append(nodeMap[cardTypeStr], nodeName)
	}

	return nodeMap, nil
}
