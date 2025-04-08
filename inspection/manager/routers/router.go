// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/beego/beego/v2/server/web"

	"infrahi/backend/inspection-manager/controllers"
)

func init() {
	ns := web.NewNamespace("/inspection/api/v1",
		web.NSRouter("/create_job", &controllers.JobController{}, "post:Post"),
		web.NSRouter("/test_result_info", &controllers.JobController{}, "get:GetAll"),
		web.NSRouter("/job/retry", &controllers.JobController{}, "post:Retry"),
		web.NSRouter("/job/detail", &controllers.JobController{}, "get:GetDetail"),

		web.NSRouter("/result/summary", &controllers.TestInfoController{}, "get:Get"),
		web.NSRouter("/result/detail", &controllers.TestInfoController{}, "get:GetOne"),
		web.NSRouter("/test_result_details", &controllers.TestInfoController{}, "get:GetOld"),
		web.NSRouter("/test_case_detail", &controllers.TestInfoController{}, "get:GetOneOld"),
		web.NSRouter("/download_log", &controllers.TestInfoController{}, "post:Post"),

		web.NSRouter("/get_all_cases", &controllers.ScriptController{}, "get:Get"),
		web.NSRouter("/script/get_config", &controllers.ScriptController{}, "get:GetConfig"),
		web.NSRouter("/script/set_config", &controllers.ScriptController{}, "post:SetConfig"),

		web.NSRouter("/get_nodes_from_cluster", &controllers.NodeController{}, "get:GetAllOld"),
		web.NSRouter("/nodes", &controllers.NodeController{}, "get:GetAll"),
		web.NSRouter("/node/resource", &controllers.NodeController{}, "get:Resource"),
		web.NSRouter("/node/download_excel", &controllers.NodeController{}, "get:DownloadExcel"),
		web.NSRouter("/node/upload_excel", &controllers.NodeController{}, "post:UploadExcel"),

		web.NSRouter("/create_template", &controllers.TemplateController{}, "post:Post"),
		web.NSRouter("/modify_template", &controllers.TemplateController{}, "post:Patch"),
		web.NSRouter("/get_template_info", &controllers.TemplateController{}, "get:Get"),
		web.NSRouter("/get_all_templates", &controllers.TemplateController{}, "get:GetAll"),
		web.NSRouter("/get_template_detail", &controllers.TemplateController{}, "get:GetOne"),
		web.NSRouter("/delete_template", &controllers.TemplateController{}, "delete:Delete"),
	)
	web.AddNamespace(ns)
}
