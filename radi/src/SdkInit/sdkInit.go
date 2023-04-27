package SdkInit

import (
	"fmt"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
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
}

func Sdkinit(sdkinfo SdkInfo, resClient *ResClient) error{
	//get config provider from config.yml file
	configProvider := config.FromFile(sdkinfo.ConfigPath)

	//init a new fabric Instance
	fabsdks, err := fabsdk.New(configProvider)

	//judge err
	if err != nil{
		fmt.Println("error in fabric-sdk init!: ", err)
		return err
	}else {
		resClient.Fbsdk = fabsdks
		fmt.Println("successfully init the fabric-sdk")
		return nil
	}
}

func ResMgmtClientCreate(resClient *ResClient, sdkinfo SdkInfo) error{
	//clientContext allows creation of transactions using the supplied identity as the credential
	clientContext := resClient.Fbsdk.Context(fabsdk.WithUser(sdkinfo.OrgAdmin), fabsdk.WithOrg(sdkinfo.OrgName))

	// Resource management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case Radi Org admin)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil{
		fmt.Println("error in resMgmtClient creation: ", err)
		return err
	}else{
		resClient.ResMgmtClient = resMgmtClient
		fmt.Println("successfully create the resmgmtCLient")
		return nil
	}
}

func MspClientCreate(resClient *ResClient, sdkinfo SdkInfo) error{
	//get signing identity of Admin user of Org Radi for channel create
	//e.g. mspclient is used for user management such as user register/user enroller/get signature
	mspClient, err := mspclient.New(resClient.Fbsdk.Context(), mspclient.WithOrg(sdkinfo.OrgName))
	if err != nil{
		fmt.Println("err in mspClient creation: ", err)
		return err
	}else{
		resClient.MspClient = mspClient
		fmt.Println("successfully create the mspClient")
		return nil
	}
}

func ChannelCreate(sdkinfo SdkInfo, resClient *ResClient) string{

//get res management client
	err := ResMgmtClientCreate(resClient, sdkinfo)
	if err != nil{
		return ""
	}
//get msp client
	err = MspClientCreate(resClient, sdkinfo)
	if err != nil{
		return ""
	}
//get signing Identity	by msp client
	signingIdentity, err := resClient.MspClient.GetSigningIdentity(sdkinfo.OrgAdmin)
	if err != nil{
		fmt.Println("err in get signing identity: ", err)
		return ""
	}

//construct channel creation request and send channel creation tx by saveChannel func
	channelReq := resmgmt.SaveChannelRequest{
		ChannelID:         sdkinfo.ChannelId,
		ChannelConfigPath: sdkinfo.ChannelConfigPath,
		SigningIdentities: []msp.SigningIdentity{signingIdentity},
	}

//send savechannel request to orderers
	saveChannelResponse, err := resClient.ResMgmtClient.SaveChannel(channelReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(sdkinfo.OrdererEndPoint))
	if err != nil{
		fmt.Println("err in channel creation : ", err)
	}else{
		fmt.Println("successfully creating CHANNEL : " + sdkinfo.ChannelId)
	}

//e.g. type TransactionID string
/*type saveChannelResponse struct{
		TransactionId
		}*/
	return string(saveChannelResponse.TransactionID)

}

//peer join channel
func JoinChannel (info SdkInfo, client *ResClient) error{
	err := client.ResMgmtClient.JoinChannel(info.ChannelId, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(info.OrdererEndPoint))
	if err != nil{
		fmt.Println("err in peers of Org Radi to join CHANNEL " + info.ChannelId + " : ", err)
		return err
	}else{
		fmt.Println("successfully make peers of Org " + info.OrgName + " to join CHANNEL : " + info.ChannelId)
		return nil
	}
}