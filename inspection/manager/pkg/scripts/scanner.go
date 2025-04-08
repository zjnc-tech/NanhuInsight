package scripts

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	beego "github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/models"
)

func ScanScript() {
	// 获取脚本路径配置
	path, err := getScriptPath()
	if err != nil {
		log.Fatalf("Error getting the script path %v: %v", path, err)
		return
	}

	// 遍历文件路径
	err = filepath.Walk(path, processFile)
	if err != nil {
		log.Printf("Error walking the path %v: %v", path, err)
		return
	}

	return
}

// getScriptPath 获取脚本路径配置
func getScriptPath() (string, error) {
	return beego.AppConfig.String(beego.BConfig.RunMode + "::script_path")
}

// processFile 处理每个文件
func processFile(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !info.IsDir() && filepath.Ext(path) == ".json" {
		updateScriptFromJSON(path)
	}
	return nil
}

// updateScriptFromJSON 更新数据库记录
func updateScriptFromJSON(filePath string) {
	// 读取文件内容
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %v: %v", filePath, err)
		return
	}

	// 解析 JSON 数据
	var script models.Script
	err = json.Unmarshal(file, &script)
	if err != nil {
		log.Printf("Error parsing file %v: %v", filePath, err)
		log.Printf("File content: %v", file)
		return
	}

	err = models.AddScript(&script)
	if err != nil {
		log.Printf("Error adding script %v: %v", script, err)
		return
	}

	return
}
