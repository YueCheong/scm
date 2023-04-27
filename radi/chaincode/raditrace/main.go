package main
import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

)
//定义链码实现结构体 实现Chaincode接口
type CCStrcut struct {

}

//Init method of Chaincode interface
func (t *CCStrcut) Init(stub shim.ChaincodeStubInterface) pb.Response{

	//初始化UsrIdCOnst、DirIdConst、DataFileIdConst并存入世界状态账本
	ledgerIdConst := IdConst{
		UserIdConst:   1,
		DirIdConst:    1,
		DataFileConst: 1,
		LogConst:      1,
	}
	bytes, err := json.Marshal(ledgerIdConst)
	if err != nil{
		return shim.Error("Err in Init Func : LedgerIdConst 序列化失败！")
	}
	err = stub.PutState("ledgerIdConst", bytes)
	if err != nil{
		return shim.Error("Err in Init Func : LedgerIdConst 存入状态账本失败！")
	}

	//初始化根目录并存储至账本上
	timeStamp, _ := GetTimeAsTemplate(stub)
	root := &Dir{
		AbsolutePath: "/dir",
		ParentDir:    "/",
		DirName:      "dir",
		DirId:        "dir_root",
		Type:         DirType,
		Size:         0,
		Content:      nil,
		Creator:      "CA",
		ModifyTime:   timeStamp,
		Remark:       "The Root Dir",
		Attributes:   "",
		TxId:         stub.GetTxID(),
	}
	rootJson, _ := json.Marshal(root)
	_ = stub.PutState(root.AbsolutePath, rootJson)

	return shim.Success([]byte("Init Func Success!"))
}

//Invoke method of Chaincode interface
func (t *CCStrcut) Invoke(stub shim.ChaincodeStubInterface) pb.Response{

	//get funcName and args for Invoke
	funcName, args := stub.GetFunctionAndParameters()

	//select method to execute
	if funcName == "DirRegister" {
		return DirRegister(stub, args)
	} else if funcName == "UsrRegister" {
		return UsrRegister(stub, args)
	} else if funcName == "DataFileRegister" {
		return DataFileRegister(stub, args)
	} else if funcName == "UsrOrgAlter" {
		return UsrOrgAlter(stub, args)
	} else if funcName == "QueryByFileName" {
		return QueryByFileName(stub, args)
	} else if funcName == "QueryByCreator" {
		return QueryByCreator(stub, args)
	} else if funcName == "GetLog" {
		return GetLog(stub, args)
	} else if funcName == "DirNameUpdate" {
		return DirNameAlter(stub, args)
	} else if funcName == "DirView" {
		return DirView(stub, args)
	} else {
		respond := RespondConstruct(FuncNameUndefined, "No such FuncName Defined", nil)
		return shim.Error(string(respond))
	}

}


func main() {
	err := shim.Start(new(CCStrcut))
	if err != nil {
		fmt.Printf("Error Starting the CC Dir : %s", err)
	}
}

