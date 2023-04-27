package main

import (
	"github.com/astaxie/beego/logs"
	//_ "radiProvenance/envSet"
	_ "radiProvenance/routers"
	_ "radiProvenance/models"
	"github.com/astaxie/beego"
)

func main() {
	err := beego.AddFuncMap("ShowAllowed", ShowAllowed)
	if err != nil {
		logs.Error("视图函数映射出错：", err.Error())
		return
	}
	beego.Run()
}


func ShowAllowed(delFlag string) string{

	if delFlag == "0" {
		return "允许下载"
	} else {
		return "禁止下载"
	}
}
