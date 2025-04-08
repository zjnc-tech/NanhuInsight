package scripts

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	beego "github.com/beego/beego/v2/server/web"
)

var mu sync.Mutex

func ZipScriptFolder(caseName string) (string, error) {
	mu.Lock() // 锁定，防止并发访问文件
	defer mu.Unlock()
	
	scriptPath, err := beego.AppConfig.String(beego.BConfig.RunMode + "::script_path")
	if err != nil {
		return "", err
	}
	// 拼接文件夹路径
	folderPath := filepath.Join(scriptPath, caseName)

	// 确保指定的文件夹存在
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return "", errors.New("指定的文件夹路径不存在: " + folderPath)
	}

	zipName := folderPath + ".zip"

	// 创建 ZIP 文件
	zipFile, err := os.Create(zipName)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历文件夹中的所有文件和子文件夹
	err = filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 创建文件的 ZIP 头信息
		relativePath, err := filepath.Rel(folderPath, filePath)
		if err != nil {
			return err
		}

		writer, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		// 打开文件并写入到 ZIP 文件中
		fileToZip, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(writer, fileToZip)
		return err
	})

	if err != nil {
		return "", err
	}

	return zipName, nil
}
