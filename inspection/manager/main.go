package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	beectx "github.com/beego/beego/v2/server/web/context"
	_ "gorm.io/driver/mysql"

	"infrahi/backend/inspection-manager/models"
	"infrahi/backend/inspection-manager/pkg/cronjob"
	"infrahi/backend/inspection-manager/pkg/scripts"
	"infrahi/backend/inspection-manager/pkg/task"
	_ "infrahi/backend/inspection-manager/pkg/utils"
	_ "infrahi/backend/inspection-manager/routers"
)

func initDB(mode string) {
	// 从配置文件中读取数据库配置
	dbURL, e1 := beego.AppConfig.String(mode + "::mysql_urls")
	dbName, e2 := beego.AppConfig.String(mode + "::mysql_db")
	dbUser, e3 := beego.AppConfig.String(mode + "::mysql_user")
	dbPass, e4 := beego.AppConfig.String(mode + "::mysql_pass")
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		log.Fatalf("配置解析失败: mysql_urls:%v, mysql_db:%v, mysql_user:%v, mysql_pass:%v, from %s",
			e1, e2, e3, e4, "NewMysqlOptions")
	}

	// 构建 MySQL 连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbUser, dbPass, dbURL, dbName)

	// 注册数据库
	err := orm.RegisterDataBase("default", "mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to register database: %v", err)
	}

	models.InitJobInfo()
	models.InitScript()
	models.InitTemplate()
	models.InitTestInfo()
	models.InitScriptConfig()

	err = orm.RunSyncdb("default", false, false)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

func migrateDatabase() error {
	o := orm.NewOrm()

	// 执行删除字段 SQL
	_, err := o.Raw(`ALTER TABLE template DROP COLUMN cron_id;`).Exec()
	if err != nil {
		log.Println("Error dropping column cron_id:", err)
		return err
	}

	return nil
}

func initLogger() {
	logPath, err := beego.AppConfig.String(beego.BConfig.RunMode + "::log_path")
	if err != nil {
		log.Fatalf("Failed to get log path: %v", err)
	}

	// 拼接完整日志文件路径
	logFilePath := fmt.Sprintf("%s/%s", logPath, "inspection-manager.log")

	// 打开日志文件，如果文件超过指定大小则进行轮转
	rotateLogFile(logFilePath)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

const maxLogSize = 10 * 1024 * 1024 // 10 MB

func rotateLogFile(logFilePath string) {
	// 检查日志文件的大小
	fileInfo, err := os.Stat(logFilePath)
	if err == nil && fileInfo.Size() > maxLogSize {
		// 如果日志文件大于 maxLogSize，进行文件轮转
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		// 将当前日志文件重命名并添加时间戳后缀
		rotatedLogPath := fmt.Sprintf("%s.%s", logFilePath, timestamp)
		err := os.Rename(logFilePath, rotatedLogPath)
		if err != nil {
			log.Fatalf("Failed to rotate log file: %v", err)
		}
	}

	// 打开日志文件，继续写入
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// 设置日志输出到文件
	log.SetOutput(file)
}

// CORS 跨域配置中间件
func CORS() beego.FilterFunc {
	return func(ctx *beectx.Context) {
		ctx.Output.Header("Access-Control-Allow-Origin", "*")
		ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-PortalID, X-UserName, X-cluster")
		ctx.Output.Header("Access-Control-Max-Age", "86400")

		// 对于 OPTIONS 请求方法，提前返回
		if ctx.Input.Method() == "OPTIONS" {
			ctx.ResponseWriter.WriteHeader(200)
		}
	}
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	if beego.BConfig.RunMode != "dev" {
		err := beego.LoadAppConfig("ini", "/opt/conf/app.conf")
		if err != nil {
			log.Printf("Failed to load app config: %v", err)
			return
		}
	}

	beego.InsertFilter("*", beego.BeforeRouter, CORS())

	initLogger()
	initDB(beego.BConfig.RunMode)
	scripts.ScanScript()
	go scripts.InitWatcher()
	cronjob.SyncCron()

	go func() {
		task.WatchJobTasks()
	}()

	beego.Run()
}
