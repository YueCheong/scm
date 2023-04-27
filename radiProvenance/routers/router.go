package routers

import (
	"github.com/astaxie/beego/context"
	"radiProvenance/controllers"
	"github.com/astaxie/beego"
)

func init() {

	//session判断
	beego.InsertFilter("/trace/*", beego.BeforeRouter, filterFunc)

	//登录
    beego.Router("/login", &controllers.UserController{}, "get:LoginShow;post:LoginHandle")

    //注册
    beego.Router("/register", &controllers.UserController{}, "get:RegisterShow;post:RegisterHandle")

	//404 Not Found
	beego.Router("/errHandle", &controllers.UserController{}, "get:ErrHandle")

	//数据集列表
    beego.Router("/trace/showMetas", &controllers.TraceController{}, "get:ShowMetas")

    //数据集上传
    beego.Router("/trace/registerDataSet", &controllers.TraceController{}, "get:RegisterDataSetShow;post:RegisterDataSet")

	//展示用户自己的数据集
	beego.Router("/trace/showMyMetas", &controllers.TraceController{}, "get:ShowMyMetas")

	//展示用户自己数据集的详情页
	beego.Router("/trace/showMyDetails", &controllers.TraceController{}, "get:ShowMyDetails")

	//游客身份查案数据集详情页
	beego.Router("/trace/showDetails", &controllers.TraceController{}, "get:ShowAllDetails")

	//下载数据集
	beego.Router("/trace/download", &controllers.TraceController{}, "post:Download")

	//下载状态管理
	beego.Router("/trace/delManage", &controllers.TraceController{}, "get:DelFlagConvert")

	//数据集内容更新
	beego.Router("/trace/alterDataSet", &controllers.TraceController{}, "get:ShowAlter;post:AlterPost")

	//日志溯源
	beego.Router("/trace/logs", &controllers.TraceController{}, "get:ShowLogs")

	//水印添加
	beego.Router("/trace/waterMarkadd", &controllers.TraceController{}, "get:ShowWaterMarkAdd;post:WaterMarkAddPost")

	//水印检测
	beego.Router("/trace/waterMarkDetect", &controllers.TraceController{}, "post:WaterMarkDetect")

	//退出登录
	beego.Router("/logout", &controllers.UserController{}, "get:Logout")
}


//过滤器函数 session判断
var filterFunc = func(ctx *context.Context) {

	session := ctx.Input.Session("userName")
	if session == nil {
		ctx.Redirect(302, "/login")
	}
}
