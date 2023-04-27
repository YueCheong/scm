package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)


//Init method of Chaincode interface
func (t *CCStruct) Init(stub shim.ChaincodeStubInterface) pb.Response{


	//初始化UsrIdCOnst、DirIdConst、DataFileIdConst并存入世界状态账本
	ledgerIdConst := IdConst{
		LogConst:      1,
	}
	bytes, err := json.Marshal(ledgerIdConst)
	if err != nil{
		return shim.Error("Err in Init Func : LedgerIdConst 序列化失败！")
	}
	err = stub.PutState(ConstIdKey, bytes)
	if err != nil{
		return shim.Error("Err in Init Func : LedgerIdConst 存入状态账本失败！")
	}

	return shim.Success([]byte("Init Func Success!"))
}

//Invoke method of Chaincode interface
func (t *CCStruct) Invoke(stub shim.ChaincodeStubInterface) pb.Response{

	//get funcName and args for Invoke
	funcName, args := stub.GetFunctionAndParameters()

	//select method to execute
	if funcName == "MetaRegister" {
		return DataRegister(stub, args)
	} else if funcName == "DataDownload" {
		return DownLoad(stub, args)
	} else if funcName == "MetaAlter" {
		return MetaAlter(stub, args)
	} else if funcName == "ShowAll" {
		return ShowAllMetaInfo(stub)
	} else if funcName == "ShowByOwner" {
		return ShowMetaByOwner(stub, args)
	} else if funcName == "ShowByDataName" {
		return ShowMetaByName(stub, args)
	} else if funcName == "ShowLogsById" {
		return ShowLogsByDataId(stub, args)
	} else if funcName == "DelData" {
		return DelMeta(stub, args)
	} else if funcName == "ShowMetaById" {
		return ShowMetaById(stub, args)
	}else {
		respond := RespondConstruct(FuncNameUndefined, "No such FuncName Defined", nil)
		return shim.Error(string(respond))
	}

}


func main() {
	err := shim.Start(new(CCStruct))
	if err != nil {
		fmt.Printf("Error Starting the CC Chaincode : %s", err)
	}
}


