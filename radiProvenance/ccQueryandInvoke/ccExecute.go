package ccQueryandInvoke

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"radiProvenance/SdkInit"
	"time"
)

func CCExecute(resclient *SdkInit.ResClient,  eventFilter string, ccId string, funId string, args []string) (channel.Response, error){

	//register a cc event for the execution
	reg, eventChan, err := resclient.ClientChannel.RegisterChaincodeEvent(ccId, eventFilter)
	if err != nil{
		fmt.Println("Err in CC Event Registration : ", ccId, " + ", eventFilter, " ", err)
		return channel.Response{}, err
	}else {
		fmt.Println("Successfully Register CC Event : ", ccId, " + ", eventFilter)
	}
	defer resclient.ClientChannel.UnregisterChaincodeEvent(reg)

	//构造传入参数
	var ccArgs [][]byte
	for _, arg := range args {
		ccArgs = append(ccArgs, []byte(arg))
	}
	ccArgs = append(ccArgs, []byte(eventFilter))


	//构造链码请求
	//execute chaincode
	//firstly create execute request
	channelReq := channel.Request{
		ChaincodeID:    ccId,
		Fcn:            funId,
		Args:           ccArgs,
	}

	//send request for execution
	response, err := resclient.ClientChannel.Execute(channelReq)
	if err != nil{
		fmt.Println("Err in ChaincodeExecution : ", err)
		return channel.Response{}, err
	}else {
		fmt.Println("Successfully Send Tx for CC Execution!")
	}

	//listen to the event and receive the response
	select {
		case ccEvent := <-eventChan :
			fmt.Println("The Tx listened by event ", ccEvent, " has been committed into Blockchain FileSystem and Updated the State Ledger!")
			return response, nil
		case <- time.After(time.Second*50) :
			fmt.Println("Err in ccEvent Receive!")
			return channel.Response{}, errors.New("")
	}
}

func QueryCC(resclient *SdkInit.ResClient, ccId string, funId string, args []string)(channel.Response, error){
	//构造传入参数
	var ccArgs [][]byte
	for _, arg := range args {
		ccArgs = append(ccArgs, []byte(arg))
	}

	channelReq := channel.Request{
		ChaincodeID:     ccId,
		Fcn:             funId,
		Args:            ccArgs,
	}

	//Query
	response, err := resclient.ClientChannel.Query(channelReq)
	if err != nil{
		fmt.Println("Err in Query CC: ", err)
		return channel.Response{}, err
	}else {
		fmt.Println("Successfully Send Tx to Query and Get Response!")
		return response, nil
	}
}