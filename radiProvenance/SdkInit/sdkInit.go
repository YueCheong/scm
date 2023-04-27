package SdkInit

import (
	"github.com/astaxie/beego/logs"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// describe the essential info of initialization of fabsdk for res management
type SdkInfo struct{
	ChannelId string			//channel Id must correspond to the channel ID in channel.tx materials

	OrgName string				//select one OrgName defined in config.yaml
	OrgAdmin string				//Admin user of the Org

	OrdererOrgName string		//OrdererOrgName defined in config.yaml
	OrdererOrgAdmin string		//Admin User of the Org
	OrdererEndPoint string		//the Orderer's host which the sdk client will send txs to

	ConfigPath string			//the path of sdk configuration: config.yaml

	ChannelConfigPath string	//the path to channel.tx materials which will be used in channel creation process
}

// record the res management client for channel createion/chaincode install/ext..
type ResClient struct {

	//fbsdk reads configuration of config.yaml file to provide context to create  res mgmt clients
	Fbsdk *fabsdk.FabricSDK
	//resMgmtClient is used for creating or updating channel
	ResMgmtClient *resmgmt.Client
	//mspClient is used for user identity management such as register/enrollment/provide sigining identity/ext..
	MspClient *mspclient.Client
	//ClientChannel is used for cc execute / query / eventhub register and receive.. ext within one channel
	ClientChannel *channel.Client
}

func Sdkinit(sdkinfo SdkInfo, resClient *ResClient) error{
	//get config provider from config.yml file
	configProvider := config.FromFile(sdkinfo.ConfigPath)

	//init a new fabric Instance
	fabsdks, err := fabsdk.New(configProvider)

	//judge err
	if err != nil{
		return err
	}else {
		resClient.Fbsdk = fabsdks
		logs.Info("successfully init the fabric-sdk")
		return nil
	}
}


//ChannelClientCreate() func is used to execute cc of the channel such as invoke/query/eventhub..ext
func ChannelClientCreate (resClient *ResClient, sdkInfo SdkInfo) error{
	
	//get the channel client creation context from sdk configuration instance
	clientChannelContext := resClient.Fbsdk.ChannelContext(sdkInfo.ChannelId, fabsdk.WithOrg(sdkInfo.OrgName), fabsdk.WithUser(sdkInfo.OrgAdmin))
	
	//New a channel client for cc execute/query and eventhub manage
	client, err := channel.New(clientChannelContext)
	if err != nil{
		return err
	}else {
		logs.Info("Successfully Create the ChannelClient Instance!")
	}
	
	resClient.ClientChannel = client
	return nil
}