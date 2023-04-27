package envSet

import (
	"github.com/astaxie/beego/logs"
	"os"
)

func init() {
	err := os.Setenv("ProjectDir", "/home/sheyufeng/data/goproject/radi")
	if err != nil{
		logs.Error("err in EnvSet: ", err)
	}
}