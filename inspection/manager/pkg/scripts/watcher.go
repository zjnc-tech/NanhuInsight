package scripts

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"infrahi/backend/inspection-manager/models"
)

var (
	watcher    *fsnotify.Watcher
	watchMutex sync.Mutex
)

func InitWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("error creating watcher: %s\n", err)
		return
	}
	defer watcher.Close()

	path, err := getScriptPath()
	if err != nil {
		log.Printf("Error get script path: %v", err)
		return
	}

	// 初始化监控
	err = watchDirectory(path)
	if err != nil {
		log.Printf("Error initializing watcher: %v", err)
	}

	// 监听文件事件
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create != 0 {
					// 检查新建的路径是文件还是目录
					info, err := os.Stat(event.Name)
					if err != nil {
						log.Printf("Error accessing path %s: %v\n", event.Name, err)
						continue
					}

					if info.IsDir() {
						// 如果是目录，递归监控该目录
						log.Printf("New directory created: %s\n", event.Name)
						err := watchDirectory(event.Name) // 调用之前的 watchDirectory 方法
						if err != nil {
							log.Printf("Error watching new directory %s: %v\n", event.Name, err)
						}
					} else if strings.HasSuffix(event.Name, ".json") {
						// 如果是 JSON 文件
						log.Printf("New JSON file created: %s\n", event.Name)
						updateScriptFromJSON(event.Name)
					}
				} else if strings.HasSuffix(event.Name, ".json") {
					// 针对 JSON 文件的其他操作
					switch {
					case event.Op&fsnotify.Write != 0:
						log.Printf("File modified: %s\n", event.Name)
						updateScriptFromJSON(event.Name)

					case event.Op&fsnotify.Remove != 0:
						log.Printf("File deleted: %s\n", event.Name)
						deleteFromDatabase(event.Name)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// 阻止主线程退出
	select {}
}

// watchDirectory 递归添加子目录及其 JSON 文件到监控中
func watchDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		// 仅监控文件夹和 JSON 文件
		if info.IsDir() {
			watchMutex.Lock()
			defer watchMutex.Unlock()
			err := watcher.Add(path)
			if err != nil {
				log.Printf("Error watching directory %s: %v\n", path, err)
			} else {
				log.Printf("Watching directory: %s\n", path)
			}
		}
		return nil
	})
}

// deleteFromDatabase 从数据库中删除记录
func deleteFromDatabase(filePath string) {
	dirName := filepath.Base(filepath.Dir(filePath))
	log.Printf("Deleting script from database: %s\n", dirName)

	err := models.DeleteScript(dirName)
	if err != nil {
		log.Printf("Error deleting script %s: %v\n", dirName, err)
	}
}
