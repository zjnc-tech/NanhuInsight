package cluster

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"infrahi/backend/inspection-manager/pkg/rpc"
	"infrahi/backend/inspection-manager/pkg/utils"
)

var ResourceMap = map[string]string{
	"Z1120":  "nvidia.com/gpu",
	"Z3200":  "nvidia.com/gpu",
	"Z2120":  "nvidia.com/gpu",
	"V5000":  "kunlunxin.com/xpu",
	"U2000":  "",
	"W64":    "metax-tech.com/gpu",
	"X10000": "mthreads.com/gpu",
}

var (
	clientSetMap = make(map[string]*kubernetes.Clientset)
	mutex        = sync.Mutex{}
)

// GetClientSet 获取 Kubernetes clientSet，如果已存在则直接返回，否则创建并存入 map
func GetClientSet(clusterName string) (*kubernetes.Clientset, error) {
	// 加锁保证并发安全
	mutex.Lock()
	defer mutex.Unlock()

	// 如果已经存在，直接返回
	if clientSet, exists := clientSetMap[clusterName]; exists {
		return clientSet, nil
	}

	// 获取 kubeconfig 配置路径
	configPath, err := utils.GetClusterConfig(clusterName)
	if err != nil {
		return nil, err
	}

	// 解析 kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	if clusterName == "x10000_prod_01" {
		config.TLSClientConfig.CAFile = ""
		config.TLSClientConfig.Insecure = true
	}

	// 创建新的 Kubernetes clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 存入 map
	clientSetMap[clusterName] = clientSet
	return clientSet, nil
}

func getNodeFreeGPU(client *kubernetes.Clientset, resource string) (map[string]bool, error) {
	var resourceName = ResourceMap[resource]
	// 获取所有命名空间中的 Pod
	pods, err := client.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %w", err)
	}

	// 记录有GPU请求的Pod所在的节点名
	gpuPodNodes := make(map[string]bool)
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			for _, container := range pod.Spec.Containers {
				gpuRequest, found := container.Resources.Requests[v1.ResourceName(resourceName)]
				if found && !gpuRequest.IsZero() {
					// 记录节点名
					gpuPodNodes[pod.Spec.NodeName] = true
					break
				}
			}
		}
	}

	return gpuPodNodes, nil
}

// GetNodeStatusMap 返回map的key为节点状态，value为节点IP列表
func GetNodeStatusMap(clusterName string, mode string, resource string) (map[string][]string, error) {
	clientSet, err := GetClientSet(clusterName)
	if err != nil {
		return nil, err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing nodes: %v", err)
		return nil, err
	}

	var gpuPodNodes map[string]bool
	if mode == "deep" && resource != "" {
		if gpuPodNodes, err = getNodeFreeGPU(clientSet, resource); err != nil {
			return nil, err
		}
	}

	result := map[string][]string{
		"master":        {},
		"ready":         {},
		"notReady":      {},
		"unschedulable": {},
		"allocated":     {},
	}

	for _, node := range nodes.Items {
		labels := node.Labels
		nodeIP := ""
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				nodeIP = address.Address
				break
			}
		}

		// 检查 Master / Control-Plane
		_, isControlPlane := labels["node-role.kubernetes.io/control-plane"]
		_, isMaster := labels["node-role.kubernetes.io/master"]
		if isControlPlane || isMaster {
			result["master"] = append(result["master"], nodeIP)
			continue
		}

		// 检查是否是 NotReady 节点
		isNotReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
				isNotReady = true
				break
			}
		}

		// 检查是否不可调度
		unschedulable := false
		for _, taint := range node.Spec.Taints {
			if taint.Effect == v1.TaintEffectNoSchedule || taint.Effect == v1.TaintEffectNoExecute {
				unschedulable = true
				break
			}
		}

		allocated := mode == "deep" && gpuPodNodes[node.Name]

		if isNotReady {
			result["notReady"] = append(result["unschedulable"], nodeIP)
		} else if unschedulable {
			result["unschedulable"] = append(result["unschedulable"], nodeIP)
		} else if allocated {
			result["allocated"] = append(result["allocated"], nodeIP)
		} else {
			result["ready"] = append(result["ready"], nodeIP)
		}
	}

	return result, nil
}

// GetNodeMapByCluster 返回map的key为节点IP，value为节点名
func GetNodeMapByCluster(clusterName string, mode string, resource string) (nodeMap map[string]string, err error) {
	clientSet, err := GetClientSet(clusterName)
	if err != nil {
		log.Printf("Error getting Kubernetes client: %v", err)
		return nil, err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing nodes: %v", err)
		return nil, err
	}

	gpuPodNodes := make(map[string]bool)
	if mode == "deep" && resource != "" {
		gpuPodNodes, err = getNodeFreeGPU(clientSet, resource)
		if err != nil {
			return nil, err
		}
	}

	nodeMap = make(map[string]string)
	for _, node := range nodes.Items {
		labels := node.Labels

		// 检查标签
		_, isControlPlane := labels["node-role.kubernetes.io/control-plane"]
		_, isMaster := labels["node-role.kubernetes.io/master"]

		// 如果标签存在，跳过节点
		if isControlPlane || isMaster {
			continue
		}

		//taints := node.Spec.Taints
		//taintDisabled := false
		//for _, taint := range taints {
		//	if taint.Key == "node.kubernetes.io/unschedulable" {
		//		taintDisabled = true
		//		break
		//	}
		//}
		//if taintDisabled {
		//	continue
		//}

		//if mode == "deep" {
		//	if gpuPodNodes[node.Name] {
		//		continue
		//	}
		//}

		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				// 深度检查且资源占用
				if mode == "deep" && gpuPodNodes[node.Name] {
					nodeMap[address.Address] = node.Name + "-skip"
				} else {
					nodeMap[address.Address] = node.Name
				}
			}
		}
	}
	log.Printf(" listing nodes of cluster %s count: %d", clusterName, len(nodeMap))

	return nodeMap, nil
}

func GetAllNodesList(clusterName string) (nodeList []string, err error) {
	clientSet, err := GetClientSet(clusterName)
	if err != nil {
		log.Printf("Error creating Kubernetes client: %v", err)
		return nil, err
	}

	nodesClient := clientSet.CoreV1().Nodes()

	nodes, listErr := nodesClient.List(context.TODO(), metav1.ListOptions{})
	if listErr != nil {
		log.Printf("Error listing nodes: %v", err)
		return nil, listErr
	}

	for _, node := range nodes.Items {
		labels := node.Labels

		// 检查标签
		_, isControlPlane := labels["node-role.kubernetes.io/control-plane"]
		_, isMaster := labels["node-role.kubernetes.io/master"]

		// 如果标签存在，跳过节点
		if isControlPlane || isMaster {
			continue
		}

		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				nodeList = append(nodeList, address.Address)
			}
		}
	}

	log.Printf(" listing nodes of cluster %s count all: %d", clusterName, len(nodeList))

	return nodeList, nil
}

// GetResourceMap 返回map的key为资源名称，value为节点IP列表
func GetResourceMap(clusterName string) (map[string][]string, error) {
	IPList, err := GetAllNodesList(clusterName)
	if err != nil {
		return nil, err
	}

	resourceMap, err := GetNodesFromAgent(clusterName, IPList)
	if err != nil {
		return nil, err
	}

	return resourceMap, nil
}

func GetProcessNodeInfo(clusterName string, mode string, IPListStr string, baseIP string,
	resource string) (rpc.JobNodesInfo, error) {
	nodeMap, err := GetNodeMapByCluster(clusterName, mode, resource)
	if err != nil {
		return rpc.JobNodesInfo{}, err
	}

	IPList := make([]string, 0, len(nodeMap))
	for key := range nodeMap {
		IPList = append(IPList, key)
	}

	resourceMap, err := GetNodesFromAgent(clusterName, IPList)
	if err != nil {
		return rpc.JobNodesInfo{}, err
	}

	filteredNodeMap := make(map[string]string)

	for _, ip := range resourceMap[resource] {
		if nodeName, exists := nodeMap[ip]; exists {
			filteredNodeMap[ip] = nodeName
		}
	}

	processNodes := make(map[string]string)

	if IPListStr == "" {
		processNodes = filteredNodeMap
	} else {
		jobIPList := strings.Split(IPListStr, ",")
		for _, ip := range jobIPList {
			if hostname, exists := filteredNodeMap[ip]; exists {
				processNodes[ip] = hostname
			}
		}
	}

	var baseNode map[string]string
	if baseIP != "" {
		if hostname, exists := nodeMap[baseIP]; exists {
			baseNode = map[string]string{baseIP: hostname}
		} else {
			baseNode = nil
			log.Printf("ERROR: base IP %s not found in node map", baseIP)
		}
	} else {
		baseNode = nil
	}

	result := rpc.JobNodesInfo{
		ProcessNodes: processNodes,
		BaseNode:     baseNode,
	}

	return result, nil
}

func GetNodesFromAgent(clusterName string, IPList []string) (map[string][]string, error) {
	agentAddr, err := utils.GetAgentAddress(clusterName)
	if err != nil {
		return nil, err
	}

	nodeMap, err := rpc.GetResourceFromAgent(agentAddr, IPList)
	if err != nil {
		return nil, err
	}

	return nodeMap, nil
}
