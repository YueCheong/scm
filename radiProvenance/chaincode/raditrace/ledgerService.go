package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

/*科学数据集元信息注册*/
func DataRegister(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//1. 接收参数
	if len(args) != 7 {
		respond :=RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}

	dataId := args[0]
	dataName := args[1]
	abstract := args[2]
	owner := args[3]
	hash := args[4]
	dataAddr := args[5]
	eventFilter := args[6]
	txid := stub.GetTxID()

	//2. 判断参数并进行类型转换 （后端进行 此处omit）
	//3. 构造MetaInfo结构体
	timestamps := GetTimeAsTemplate(stub)
	dataset := &MetaData{
		DataId:    dataId,
		DataName:  dataName,
		Abstract:  abstract,
		Owner:     owner,
		Hash:      hash,
		DataAddr:  dataAddr,
		TimeStamp: timestamps,
		DelFlag:   "0",
		Type:      MetaType,
		TxId:      txid,
	}

	//4. json序列化
	jsonbytes, err := json.Marshal(dataset)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//5. 构造composite key
	prikey, err := stub.CreateCompositeKey(MetaPre, []string{IdAttr, dataId})
	if err != nil {
		respond := RespondConstruct(CompositekeyError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//6. 判断主键存在性
	if IsKeyExist(stub, prikey) {//存在
		respond := RespondConstruct(KeyExisitedError, "dataId is Existed", nil)
		return shim.Error(string(respond))
	}


	//7. 数据上链
	err = stub.PutState(prikey, jsonbytes)
	if err != nil {
		respond := RespondConstruct(PutStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//8. 注册日志信息
	err = LogRegister(stub, dataId, dataName, owner, owner, "Register", timestamps, txid)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//9. 注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//10. 构造注册成功返回信息
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)
}


//展示全部元信息
func ShowAllMetaInfo(stub shim.ChaincodeStubInterface) pb.Response {
	//1 接收参数 (无需任何参数)

	//2 使用partial key进行范围查询
	iterator ,err := stub.GetStateByPartialCompositeKey(MetaPre, []string{IdAttr})
	if err != nil {
		respond := RespondConstruct(CompositekeyError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//3 迭代获取全部的MetaInfo信息

	metas := &MetaRespond{
		MetaDatas:[]MetaData{},
	}

	for iterator.HasNext() {
		item,_ := iterator.Next()
		jsonbytes := item.GetValue()

		//反序列化
		metaInfo := new(MetaData)
		err := json.Unmarshal(jsonbytes, metaInfo)
		if err != nil {
			respond := RespondConstruct(UnMarshalError, err.Error(), nil)
			return shim.Error(string(respond))
		}

		metas.MetaDatas = append(metas.MetaDatas, *metaInfo)
	}


	//4 构造成功返回信息
	jsonbytes, err := json.Marshal(metas)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)

}

//返回指定用户名的元信息集合
func ShowMetaByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//1. 接收参数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}

	owner := args[0]

	//2. 创建福查询字符串
	richStr := fmt.Sprintf(`{"selector":{"owner":"%s", "type":"%s"}}`, owner, MetaType)

	//3. counchDB富查询
	iterator, err := stub.GetQueryResult(richStr)
	if err != nil {
		respond := RespondConstruct(RichQueryError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//4. 迭代获取MetaData
	metas := &MetaRespond{
		MetaDatas: []MetaData{},
	}


	for iterator.HasNext() {

		item, _ := iterator.Next()
		jsonbytes:= item.GetValue()

		metaInfo := new(MetaData)
		err := json.Unmarshal(jsonbytes, metaInfo)

		if err != nil {
			respond :=RespondConstruct(UnMarshalError, err.Error(), nil)
			return shim.Error(string(respond))
		}

		metas.MetaDatas = append(metas.MetaDatas, *metaInfo)
	}

	//5. 构造成功返回结构体
	jsonbytes, err := json.Marshal(metas)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)

}

//返回指定数据名的元信息
func ShowMetaByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//1. 接收参数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}
	fileName := args[0]

	//2. 创建富查询字符串
	richstr := fmt.Sprintf(`{"selector":{"data_name":"%s", "type":"%s"}}`, fileName, MetaType)

	//3. counchDB富查询
	iterator, err := stub.GetQueryResult(richstr)
	if err != nil {
		respond := RespondConstruct(RichQueryError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//4. 迭代获取相应的数据集元信息
	metas := &MetaRespond{
		MetaDatas: []MetaData{},
	}

	for iterator.HasNext() {
		item, _ := iterator.Next()
		jsonbytes := item.GetValue()

		metaInfo := new(MetaData)
		err := json.Unmarshal(jsonbytes, metaInfo)
		if err != nil {
			respond :=RespondConstruct(UnMarshalError, err.Error(), nil)
			return shim.Error(string(respond))
		}

		metas.MetaDatas = append(metas.MetaDatas, *metaInfo)
	}


	//5. 构造成功返回结构体
	jsonbytes, err := json.Marshal(metas)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)


}

//修改数据集元信息
func MetaAlter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//1. 接收参数
	if len(args) != 7 {
		respond :=RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}

	dataId := args[0]
	dataName := args[1]
	abstract := args[2]
	owner := args[3]
	hash := args[4]
	dataAddr := args[5]
	eventFilter := args[6]
	txid := stub.GetTxID()

	//2. 判断参数并进行类型转换 （后端进行 此处omit）
	//3. 构造MetaInfo结构体
	timestamps := GetTimeAsTemplate(stub)
	dataset := &MetaData{
		DataId:    dataId,
		DataName:  dataName,
		Abstract:  abstract,
		Owner:     owner,
		Hash:      hash,
		DataAddr:  dataAddr,
		TimeStamp: timestamps ,
		DelFlag:   "0",
		Type:      MetaType,
		TxId:      txid,
	}

	//4. json序列化
	jsonbytes, err := json.Marshal(dataset)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//5. 构造composite key
	prikey, err := stub.CreateCompositeKey(MetaPre, []string{IdAttr, dataId})
	if err != nil {
		respond := RespondConstruct(CompositekeyError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//6. 判断主键存在性
	if !IsKeyExist(stub, prikey) {//不存在
		respond := RespondConstruct(KeyExisitedError, "dataId is not Existed", nil)
		return shim.Error(string(respond))
	}

	//7. 数据上链
	err = stub.PutState(prikey, jsonbytes)
	if err != nil {
		respond := RespondConstruct(PutStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//8. 注册日志
	err = LogRegister(stub, dataId, dataName, owner, owner, "Alter", timestamps, txid)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//9. 注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	//10. 构造注册成功返回信息
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)
}


//元数据删除
func DelMeta(stub shim.ChaincodeStubInterface, args []string) pb.Response{

	//1.接收参数
	if len(args) != 2 {
		respond :=RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}
	dataId := args[0]
	eventFilter := args[1]
	txid := stub.GetTxID()

	//2.判断是否存在
	prikey, err := stub.CreateCompositeKey(MetaPre, []string{IdAttr, dataId})
	if err != nil {
		respond := RespondConstruct(CompositekeyError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	if !IsKeyExist(stub, prikey) {//不存在
		respond := RespondConstruct(KeyExisitedError, "dataId is not Existed", nil)
		return shim.Error(string(respond))
	}

	//3.获取该MetaData struct
	jsonbytes, err := stub.GetState(prikey)
	if err != nil {
		respond := RespondConstruct(GetStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	metaInfo := new(MetaData)
	err = json.Unmarshal(jsonbytes, metaInfo)
	if err != nil {
		respond := RespondConstruct(UnMarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//4.修改DelFlag并重新上链
	timeStamps := GetTimeAsTemplate(stub)
	if metaInfo.DelFlag == "0" {
		metaInfo.DelFlag = "1"
	} else {
		metaInfo.DelFlag = "0"
	}
	metaInfo.TimeStamp = timeStamps

	jsonbytes, err = json.Marshal(metaInfo)
	err = stub.PutState(prikey, jsonbytes)
	if err != nil {
		respond := RespondConstruct(PutStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//5. 注册日志
	err = LogRegister(stub, dataId, metaInfo.DataName, metaInfo.Owner, metaInfo.Owner, "DelManage", timeStamps, txid)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//6. 注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//7. 返回成功响应结果
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)

}


//日志注册
func LogRegister(stub shim.ChaincodeStubInterface, dataId, dataName, owner, operator, operation, timestamps, txid string ) error {

	//1. 创建日志结构体变量
	loginfo := &LogInfo{
		DataId:    dataId,
		DataName:  dataName,
		Owner:     owner,
		Operator:  operator,
		Operation: operation,
		TimeStamp: timestamps,
		TxId:      txid,
		Type:      LogType,
	}

	//2. 序列化
	logJson, _ := json.Marshal(loginfo)


	//3. 创建日志的composite key
	logId, err := GetLedgerConst(stub)
	if err != nil {
		return err
	}

	logIdStr := strconv.FormatUint(logId, 10)
	prikey, err := stub.CreateCompositeKey(LogPre, []string{IdAttr, logIdStr})
	if err != nil {
		return err
	}

	//4. 存储
	err = stub.PutState(prikey, logJson)
	if err != nil {
		return err
	} else {
		err := UpdateLedgerConst(stub)
		if err != nil {
			return err
		} else{
			return nil
		}
	}
}

//下载
func DownLoad(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//1. 接收参数
	if len(args) != 5 {
		respond :=RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}
	dataId := args[0]
	dataName := args[1]
	owner := args[2]
	operator := args[3]
	eventFilter := args[4]
	txid := stub.GetTxID()

	operation := "Download"
	timestamps := GetTimeAsTemplate(stub)

	//2. 注册下载日志
	err := LogRegister(stub, dataId, dataName, owner, operator, operation, timestamps, txid)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//3. 注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//4. 返回结果
	respond := RespondConstruct(Success, SuccessMsg, nil)
	return shim.Success(respond)
}


//返回日志数据
func ShowLogsByDataId(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//1. 接收参数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}
	dataId := args[0]

	//2.创建富查询字符串
	richStr := fmt.Sprintf(`{"selector":{"data_id":"%s", "type":"%s"}}`, dataId, LogType)

	//3. counchDB富查询
	iterator, err := stub.GetQueryResult(richStr)
	if err != nil {
		respond := RespondConstruct(RichQueryError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//4. 迭代
	logs := &LogRespond{LogInfos: []LogInfo{}}

	for iterator.HasNext() {
		item,_ := iterator.Next()
		jsonbytes := item.Value

		log := new(LogInfo)
		err := json.Unmarshal(jsonbytes, log)
		if err != nil {
			respond := RespondConstruct(UnMarshalError, err.Error(), nil)
			return shim.Error(string(respond))
		}

		logs.LogInfos = append(logs.LogInfos, *log)
	}

	//返回
	jsonbytes, err := json.Marshal(logs)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	respond := RespondConstruct(Success, SuccessMsg, jsonbytes)
	return shim.Success(respond)

}

//根据数据Id获取数据集
func ShowMetaById(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//1. 接收参数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "", nil)
		return shim.Error(string(respond))
	}

	dataId := args[0]

	//2. 创建查询数据集主键
	prikey, err := stub.CreateCompositeKey(MetaPre, []string{IdAttr, dataId})
	if err != nil {
		respond := RespondConstruct(CompositekeyError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//3. 查询数据集
	bytes, err := stub.GetState(prikey)
	if err != nil {
		respond := RespondConstruct(GetStateError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//4. 构造数据集响应结构并返回
	respond := RespondConstruct(Success, SuccessMsg, bytes)
	return shim.Success(respond)
}