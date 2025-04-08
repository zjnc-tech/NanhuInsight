package utils

import (
	"fmt"
	"log"
	"net"
	"path/filepath"

	beego "github.com/beego/beego/v2/server/web"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient dynamic.Interface
)

// GetK8sClient 获取全局 Kubernetes client
func GetK8sClient() dynamic.Interface {
	return k8sClient
}

func GetManagerConfig() string {
	mode := beego.BConfig.RunMode
	path, err := beego.AppConfig.String(mode + "::kube_path")
	if err != nil {
		log.Printf("Failed to get manager kubeconfig: %v", err)
		return ""
	}
	return filepath.Join(path, mode)
}

func GetClusterConfig(clusterName string) (string, error) {
	mode := beego.BConfig.RunMode
	path, err := beego.AppConfig.String(mode + "::kube_path")
	if err != nil {
		return "", err
	}
	return filepath.Join(path, clusterName), nil
}

func GetAgentAddress(clusterName string) (string, error) {
	ip, e1 := beego.AppConfig.String(clusterName + "::agent_ip")
	port, e2 := beego.AppConfig.String(clusterName + "::agent_port")
	if e1 != nil || e2 != nil {
		return "", fmt.Errorf("配置解析失败, agent_ip:%v, agent_port:%v", e1, e2)
	}
	return net.JoinHostPort(ip, port), nil
}

func init() {
	if beego.BConfig.RunMode == "dev" {
		fmt.Println("k8s not supported in dev mode yet")
		return
	}

	configPath := GetManagerConfig()
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Printf("build config error: %v", err)
		return
	}

	k8sClient, err = dynamic.NewForConfig(k8sConfig)
	if err != nil {
		log.Printf("create client error: %v", err)
		return
	}

	log.Printf("Kubernetes client of manager cluster initialized successfully")
}
