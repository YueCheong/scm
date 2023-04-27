package main

import (
	"fmt"
	"os"
	"radi/SdkInit"
	"radi/ccmgmt"
	"radi/envSet"

)

const (
	configpath = "../sdkconfig.yaml"
)
func main(){

	err := envSet.EnvSet()
/*Set system environment*/
	if err != nil{
		return
	}


/* Init the SdkInfo struct */
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

/*Define the resClient struct to manage resoruces*/
	resClient := new(SdkInit.ResClient)

/*initialize the sdk instance*/
	err = SdkInit.Sdkinit(sdkinfo, resClient)
	if err != nil{
		return
	}

	//defer
	defer resClient.Fbsdk.Close()

/*create channel by sdk instance*/
	txid := SdkInit.ChannelCreate(sdkinfo, resClient)
	fmt.Println("CHANNEL " + sdkinfo.ChannelId + " 's creating txid is : ", txid)

/*join channel by org peers */
	err = SdkInit.JoinChannel(sdkinfo, resClient)
	if err != nil{
		return
	}

/*package & install & instantiate the cc*/
	txid = ccmgmt.InstallAndInitCC(sdkinfo, resClient)
	fmt.Println("CC" + " radiTrace " + "INSTANTIATE's txid is : ", txid)

}
