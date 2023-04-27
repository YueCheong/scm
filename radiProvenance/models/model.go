package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"os"
	"radiProvenance/SdkInit"
	_ "github.com/go-sql-driver/mysql"
)

const (
	configpath = "./sdkconfig.yaml"
)

var ResClient *SdkInit.ResClient

func init() {

	/*初始化联盟连资源管理客户端用于与联盟链账本交互*/
	//Init the SdkInfo struct
	sdkinfo := SdkInit.SdkInfo{
		ConfigPath:configpath,
		ChannelId: "radichannel",
		OrgName: "Radi",
		OrgAdmin:"Admin",
		OrdererOrgName: "OrdererOrg",
		OrdererOrgAdmin:"Admin",
		ChannelConfigPath: os.Getenv("ProjectDir") + "/channel-artifitial/radichannel.tx",
		OrdererEndPoint: "orderer.radi.trace.com",

	}

	//Define the resClient struct to manage resoruces*/
	ResClient = new(SdkInit.ResClient)

	//initialize the sdk instance*/
	err := SdkInit.Sdkinit(sdkinfo, ResClient)
	if err != nil{
		logs.Error("sdkInit失败：", err.Error())
		return
	}

	//defer
	//defer ResClient.Fbsdk.Close()

	//instante the channel client for CC Execution/Query/EventHub Manage..ext
	err = SdkInit.ChannelClientCreate(ResClient, sdkinfo)
	if err != nil{
		logs.Error("ChannelClient实例化失败：", err.Error())
		return
	}

	/*mysql建表初始化*/
	//连接数据库
	orm.RegisterDataBase("default", "mysql",
		"root:123456@tcp(127.0.0.1:3306)/casearth?charset=utf8")

	//建表
	orm.RegisterModel(new(User), new(DataSetInfo))

	//执行上述更新
	orm.RunSyncdb("default", false, true)


}
