package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"strings"
	"time"
)

//从账本获取三类资源的Id序号用于创建资源id Id序号自增后存入账本
func GetLedgerConst(stub shim.ChaincodeStubInterface, keyClass int) (uint64, error){

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
	if keyClass == UserKey {	//获取用户Id
		keyClassId = constId.UserIdConst
	} else if keyClass == DirKey {	//获取目录Id
		keyClassId = constId.DirIdConst
	} else if keyClass == DataFileKey {	//获取数据文件Id
		keyClassId = constId.DataFileConst
	} else if keyClass == LogKey {
		keyClassId =constId.LogConst
	} else {
		return 0, errors.New("No such KeyClass Defined!")
	}

	return keyClassId, nil

}


func UpdateLedgerConst(stub shim.ChaincodeStubInterface, keyClass int) error {

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

	//相应Id Const自增
	if keyClass == UserKey {	//获取用户Id
		constId.UserIdConst++
	} else if keyClass == DirKey {	//获取目录Id
		constId.DirIdConst++
	} else if keyClass == DataFileKey {	//获取数据文件Id
		constId.DataFileConst++
	} else if keyClass == LogKey {
		constId.LogConst++
	} else {
		return errors.New("No such KeyClass Defined!")
	}

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
//获取该交易请求的时间戳 并转化成年月日表示形式
func GetTimeAsTemplate(stub shim.ChaincodeStubInterface) (string, error){
	//获取交易时间戳
	txTimeStamp, err := stub.GetTxTimestamp()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err in Get TxTimeStamp : %s", err))
	}

	//获取时间 单位秒 && 毫秒  以uint64类型表示
	seconds := txTimeStamp.Seconds
	nanos := int64(txTimeStamp.Nanos)
	//将时间以模板所示类型表示 返回时间字符串
	timeTemplate := "2006-01-02 15:04:05"
	return time.Unix(seconds, nanos).Format(timeTemplate), nil

}

//构造响应结构体
func RespondConstruct(respondCode string, respondMsg string, respondData interface{}) []byte{

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


//判断键是否存在
func IsKeyExist(stub shim.ChaincodeStubInterface, key string) bool{

	bytes, err := stub.GetState(key)
	if bytes == nil || err != nil{
		return false
	} else {
		return true
	}

}

//父目录下资源数量增加（目录注册和数据资源上链使用）
//父目录的size加一 且将新增资源的资源名（DirName或者DataFileName）填入Content中
func AddSrcToParentDir(stub shim.ChaincodeStubInterface, parentDirPath string, child Child) error {

	//从账本获取ParentDir变量
	parentJson, err := stub.GetState(parentDirPath)
	if err != nil {
		return err
	}

	//转换成dir变量
	parentDir := new(Dir)
	err = json.Unmarshal(parentJson, parentDir)
	if err != nil {
		return err
	}

	//修改parentDir的content和size
	content := parentDir.Content
	content = append(content, child)
	parentDir.Content = content

	parentDir.Size ++

	//将修改后的Dir变量再次序列化 并存入数据账本中
	dirJson, err := json.Marshal(parentDir)
	if err != nil {
		return err
	}

	err = stub.PutState(parentDirPath, dirJson)
	if err != nil {
		return err
	} else {
		return nil
	}


}

//日志注册
func LogRegister(stub shim.ChaincodeStubInterface, action, decision, txid, timestamp string ) error {

	//创建日志结构体变量
	loginfo := &LogInfo{
		LogAction: action,
		Decision:  decision,
		Txid:      txid,
		TimeStamp: timestamp,
	}

	//序列化
	logJson, _ := json.Marshal(loginfo)


	//创建日志存储key
	logId, err := GetLedgerConst(stub, LogKey)
	logKey := LogPreString + strconv.FormatUint(logId, 10)
	if err != nil {
		return err
	}

	//存储
	err = stub.PutState(logKey, logJson)
	if err != nil {
		return err
	} else {
		return nil
	}
}

//先修改其孩子节点的内容 最后改自己的(算是递归后续遍历吧)
func ChangeDirName(stub shim.ChaincodeStubInterface, preAbPath string, newParentPath string, newAbPath string, level int) error {

	//获取请求修改的资源变量
	prekey := preAbPath
	newKey := newAbPath
	dirJson, _ := stub.GetState(prekey)

	dir := new(Dir)
	_ = json.Unmarshal(dirJson, dir)
	childs := dir.Content

	//修改子目录节点的parentDir和absoluteDir
	for _, child := range childs {

		preChildAbPath := preAbPath + "/" + child.SrcName
		newChildAbPath := newAbPath + "/" + child.SrcName

		if child.SrcType == "Dir" {		//子目录资源做递归更改
			err := ChangeDirName(stub, preChildAbPath, newAbPath, newChildAbPath, level+1)
			if err != nil {
				return err
			}
		} else {	//该目录下的数字文件资源则直接修改即可
			err := AlterFilePath(stub, preChildAbPath, newAbPath, newChildAbPath)
			if err != nil {
				return err
			}
		}
	}

	//修改本级目录的内容
	dir.ParentDir =newParentPath
	dir.AbsolutePath = newKey
	if level == 1 {
		dirlists := strings.Split(dir.AbsolutePath, "/")
		dir.DirName = dirlists[len(dirlists)-1]
	}

	//重新存储该目录
	dirJson, err := json.Marshal(dir)
	if err != nil {
		return err
	}

	err = stub.PutState(newKey, dirJson)
	if err != nil {
		return err
	}

	err = stub.DelState(prekey)
	return err


}

//修改某数据文件路径(DFS)
func AlterFilePath(stub shim.ChaincodeStubInterface, preAbPath string, newParentPath string, newAbPath string) error {

	dataFile := new(DataFile)
	fileJson, _ := stub.GetState(preAbPath)
	_ = json.Unmarshal(fileJson, dataFile)

	dataFile.AbsolutePath = newAbPath
	dataFile.ParentDir = newParentPath

	//重新序列化
	dataFileJson, err := json.Marshal(dataFile)
	if err != nil {
		return err
	}
	newKey := newAbPath

	err = stub.PutState(newKey, dataFileJson)
	if err != nil {
		return err
	}

	err = stub.DelState(preAbPath)
	return err
}

func DirRetConstruct(stub shim.ChaincodeStubInterface, this *DirRet, abPath string) error {

	dirJson, err := stub.GetState(abPath)
	if err != nil {
		return err
	}

	dir := new(Dir)
	err = json.Unmarshal(dirJson, dir)
	if err != nil {
		return err
	}


	childs := dir.Content
	var childDirs []string
	for _, child := range childs {
		if child.SrcType != "Dir" {
			childDataFile := DirRet{
				SrcAbPath: abPath + "/" + child.SrcName,
				SrcType:   child.SrcType,
				Childs:    nil,
			}

			this.Childs = append(this.Childs, childDataFile)
		} else {
			cdirPath := abPath + "/" + child.SrcName
			childDirs = append(childDirs, cdirPath)
		}

	}

	for _, childDir := range childDirs {
		next := new(DirRet)
		next.SrcType = "Dir"
		next.SrcAbPath = childDir
		err = DirRetConstruct(stub, next, childDir)
		if err != nil {
			return err
		}

		this.Childs = append(this.Childs, *next)
	}

	return nil

}