package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"time"
	"errors"
)

//构造响应结构体
func RespondConstruct(respondCode string, respondMsg string, respondData []byte) []byte{

	//构建Respond响应变量
	respond := &Respond{
		RespondCode: respondCode,
		RespondMsg:  respondMsg,
		RespondData: respondData,
	}

	respondJson, err:= json.Marshal(respond)

	if err != nil {
		respond.RespondCode = RespondConstructError
		respond.RespondMsg = err.Error()
		respondData = nil

		//字段固定 本次序列化不可能出错
		respondJson, _ = json.Marshal(respond)
	}

	return respondJson
}


//获取该交易请求的时间戳 并转化成年月日表示形式
func GetTimeAsTemplate(stub shim.ChaincodeStubInterface) string{
	////获取交易时间戳 不做错误判断了
	//txTimeStamp, _ := stub.GetTxTimestamp()
	//
	////获取时间 单位秒 && 毫秒  以uint64类型表示
	//seconds := txTimeStamp.Seconds
	//nanos := int64(txTimeStamp.Nanos)

	//将时间以模板所示类型表示 返回时间字符串
	timeTemplate := "2006-01-02 15:04:05"

	return time.Now().Format(timeTemplate)

}

//判断键是否存在
func IsKeyExist(stub shim.ChaincodeStubInterface, key string) bool{

	bytes, err := stub.GetState(key)
	if bytes == nil || err != nil{ //不存在
		return false
	} else { //存在
		return true
	}
}


//从账本获取三类资源的Id序号用于创建资源id Id序号自增后存入账本
func GetLedgerConst(stub shim.ChaincodeStubInterface) (uint64, error){

	//从账本中获取LedgerConst结构体
	ledgerConstjson, err := stub.GetState(ConstIdKey)
	if err != nil{
		return 0, errors.New(fmt.Sprintf("Err in  Get ConstId: %s", err))
	}

	//反序列化获得constId结构体变量
	constId := new(IdConst)
	err = json.Unmarshal(ledgerConstjson,constId)
	if err != nil{
		return 0, errors.New(fmt.Sprintf("Err in UsrRegister of unmarshal constId: %s", err))
	}

	var keyClassId uint64
	keyClassId =constId.LogConst
	return keyClassId, nil

}

func UpdateLedgerConst(stub shim.ChaincodeStubInterface) error {

	//从账本中获取LedgerConst结构体
	ledgerConstjson, err := stub.GetState(ConstIdKey)
	if err != nil{
		return errors.New(fmt.Sprintf("Err in  Get ConstId: %s", err))
	}

	//反序列化获得constId结构体变量
	constId := new(IdConst)
	err = json.Unmarshal(ledgerConstjson,constId)
	if err != nil{
		return errors.New(fmt.Sprintf("Err in UsrRegister of unmarshal constId: %s", err))
	}

	constId.LogConst++

	//序列化更新后的ConstId结构体变量 并存入Ledger
	ledgerConstjson, err = json.Marshal(constId)
	if err != nil{
		return errors.New(fmt.Sprint("Err in Marshal ledgerConst : %s", err))
	}
	err = stub.PutState(ConstIdKey, ledgerConstjson)
	if err != nil{
		return errors.New(fmt.Sprint("Err in PutState ledgerConst : %s", err))
	}

	//返回获取的IdConst用于创建UsrId/DirId/DataFileId
	return nil

}




