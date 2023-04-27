package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"

)

//用户注册 传入参数为{Org, SubName, Pubkey, Issuer, Expire, Signature, Attributes}
func UsrRegister(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) != 8{
		respond := RespondConstruct(ArgsNumberError, "Error Parameters of UsrRegister", nil)
		return shim.Error(string(respond))
	}

	//get args
	org := args[0]
	subName := args[1]
	pubKey := args[2]
	issuer := args[3]
	expire := args[4]
	signature := args[5]
	attributes := args[6]
	eventName := args[7]

	//创建usr在状态账本上存储的键
	usrKey := UsrPreString + org + subName


	//判断usr主键是否已存在
	if IsKeyExist(stub, usrKey) {
		respond := RespondConstruct(KeyExisitedError, fmt.Sprintf("The Usr Primary Key is existed"), nil)
		return shim.Error(string(respond))
	}

	//从账本上获取ConstId.UsrIdConst
	userId, err := GetLedgerConst(stub, UserKey)
	if err != nil {
		respondJson := RespondConstruct(GetConstIDError, err.Error(), nil)
		return shim.Error(string(respondJson))
	}

	//创建该usr的usrId
	subjectId := UsrPreString + strconv.FormatUint(userId, 10)

	//获取Txid
	txid := stub.GetTxID()

	//获取时间
	timestamp, _ := GetTimeAsTemplate(stub)

	//创建usr变量
	usr := &Usr{
		Org:         org,
		SubjectName: subName,
		SubjectId:   subjectId,
		Pubkey:      pubKey,
		Issuer:      issuer,
		Expire:      expire,
		Signature:   signature,
		Attributes:  attributes,
		TxId:        txid,
	}

	//序列化usr变量为json bytes
	usrJson, err := json.Marshal(usr)
	if err != nil {
		respondJson := RespondConstruct(MarshalError, fmt.Sprintf("Err in Marshal Usr : %s", err), nil)
		return shim.Error(string(respondJson))
	}

	//将usr键值对存入状态账本
	err = stub.PutState(usrKey, usrJson)
	if err != nil {
		respondJson := RespondConstruct(PutStateError, fmt.Sprintf("Err in putState Usr : %s", err), nil)
		return shim.Error(string(respondJson))
	}

	//成功注册
	respondJson := RespondConstruct(Success, SuccessMsg, usr)
	UpdateLedgerConst(stub, UserKey)

	//注册日志
	err = LogRegister(stub, UserRegisterAction, "1", txid, timestamp)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	UpdateLedgerConst(stub, LogKey)

	//注册链码事件
	err = stub.SetEvent(eventName, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	return shim.Success(respondJson)

}



//目录资源注册 传入参数为{parentDir, dirName, creator, remark, attributes}
func DirRegister(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//判断参数个数
	if len(args) != 6{
		respond := RespondConstruct(ArgsNumberError, "Error Parameters of DirRegister", nil)
		return shim.Error(string(respond))
	}

	//获取传入参数
	parentDir := args[0]
	dirName := args[1]
	creator := args[2]
	remark  := args[3]
	attributes := args[4]
	eventName := args[5]

	//获取目录绝对路径作为PrimaryKey
	absolutePath := parentDir + "/" + dirName
	dirKey := absolutePath

	//验证主键是已否存在
	if IsKeyExist(stub, absolutePath) {
		respond := RespondConstruct(KeyExisitedError, "The Dir Primary Key is Existed", nil)
		return shim.Error(string(respond))
	}

	//验证父目录是否存在
	if !IsKeyExist(stub, parentDir) {
		respond := RespondConstruct(ParentDirNotExistedError, "The ParentDir is not Existed", nil)
		return shim.Error(string(respond))
	}
	

	//获取目录Id
	dirIdConst, err := GetLedgerConst(stub, DirKey)
	if err != nil {
		respond := RespondConstruct(GetConstIDError, fmt.Sprintf("Error in Get Dir Const ID : %s", err), nil)
		return shim.Error(string(respond))
	}
	dirId := DirPreString + strconv.FormatUint(dirIdConst, 10)


	//目录下文件数目为0
	size := 0

	//目录的内容（该目录下文件的绝对路径）
	var content []Child = nil

	//创建时间
	timeStamp, err := GetTimeAsTemplate(stub)
	if err != nil {
		respond := RespondConstruct(GetTimeStampError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//获取请求交易Id
	txId := stub.GetTxID()

	//资源类型为Dir
	dirType := DirType

	//创建Dir结构体变量
	dir := &Dir{
		AbsolutePath: absolutePath,
		ParentDir:    parentDir,
		DirName:      dirName,
		DirId:        dirId,
		Type:         dirType,
		Size:         size,
		Content:      content,
		Creator:      creator,
		ModifyTime:   timeStamp,
		Remark:       remark,
		Attributes:   attributes,
		TxId:         txId,
	}


	//序列化Dir变量
	dirJson, err := json.Marshal(dir)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}


	//写入世界账本
	err = stub.PutState(dirKey, dirJson)
	if err != nil {
		respond := RespondConstruct(PutStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//成功注册Dir后
	respondJson := RespondConstruct(Success, SuccessMsg, dir)

	//更新constID
	UpdateLedgerConst(stub, DirKey)

	//对父目录进行更新
	child := Child{
		SrcName: dirName,
		SrcType: "Dir",
	}
	err = AddSrcToParentDir(stub, parentDir, child)
	if err != nil {
		respond := RespondConstruct(AddSrcToParentDirError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//注册日志
	err = LogRegister(stub, DirRegisterAction, "1", txId, timeStamp)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	UpdateLedgerConst(stub, LogKey)

	//注册链码事件
	err = stub.SetEvent(eventName, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}


	return shim.Success(respondJson)
}

//数据文件上链
func DataFileRegister(stub shim.ChaincodeStubInterface, args []string)  pb.Response{

	//判断参数个数
	if len(args) != 10 {
		respond := RespondConstruct(ArgsNumberError, "Error Parameters of DataFileRegister", nil)
		return shim.Error(string(respond))
	}

	//获取传入参数
	parentDir := args[0]
	dataFileName := args[1]
	dataFileType := args[2]
	size := args[3]
	content := args[4]
	checkSum := args[5]
	creator := args[6]
	remark := args[7]
	attributes := args[8]
	eventName := args[9]

	//获取数据文件的绝对路径作为Primary Key
	absolutePath := parentDir + "/" + dataFileName
	//验证该主键是否存在
	if IsKeyExist(stub, absolutePath) {
		respond := RespondConstruct(KeyExisitedError, "The DataFile Path is Existed!", nil)
		return shim.Error(string(respond))
	}

	//验证父目录是否存在
	if !IsKeyExist(stub, parentDir) {
		respond := RespondConstruct(ParentDirNotExistedError, "The parentDir Path is not Existed", nil)
		return shim.Error(string(respond))
	}

	//获取Id
	idConst, err := GetLedgerConst(stub, DataFileKey)
	if err != nil {
		respond := RespondConstruct(GetConstIDError, fmt.Sprintf("Error in Get DataFile Const ID : %s", err), nil)
		return shim.Error(string(respond))
	}
	DataFileId := DataFilePreString + strconv.FormatUint(idConst, 10)


	//获取交易请求时间
	timestamp, err := GetTimeAsTemplate(stub)
	if err != nil {
		respond := RespondConstruct(GetTimeStampError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//获取交易Id
	txId := stub.GetTxID()

	//创建dataFile 变量
	dataFile := &DataFile{
		AbsolutePath: absolutePath,
		ParentDir:    parentDir,
		DataFileName: dataFileName,
		DataFileId:   DataFileId,
		DataFileType: dataFileType,
		Size:         size,
		Content:      content,
		checkSum:     checkSum,
		Creator:      creator,
		ModifyTime:   timestamp,
		Remark:       remark,
		Attributes:   attributes,
		TxId:         txId,
	}

	//序列化变量
	dataFileJson, err := json.Marshal(dataFile)
	if err != nil {
		respond := RespondConstruct(MarshalError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//写入世界账本
	err = stub.PutState(absolutePath, dataFileJson)
	if err != nil {
		respond := RespondConstruct(PutStateError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//成功写入世界账本后
	respondJson := RespondConstruct(Success, SuccessMsg, dataFile)

	//更新ConstId
	UpdateLedgerConst(stub, DataFileKey)

	//更新父目录内容
	child := Child{
		SrcName: dataFileName,
		SrcType: "DataFile",
	}
	err = AddSrcToParentDir(stub, parentDir, child)
	if err != nil {
		respond := RespondConstruct(AddSrcToParentDirError, err.Error(), nil)
		return  shim.Error(string(respond))
	}

	//注册日志
	err = LogRegister(stub, DataFileRegisterAction, "1", txId, timestamp)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	UpdateLedgerConst(stub, LogKey)

	//注册链码事件
	err = stub.SetEvent(eventName, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//返回存入数据文件
	return shim.Success(respondJson)

}

//用户机构更新
//传入参数={PreOrg, SubName, NewOrg, eventFilter}
func UsrOrgAlter(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//检查参数个数
	if len(args) != 4 {
		respond := RespondConstruct(ArgsNumberError, "Error Num of Parameters of UsrOrgAlter", nil)
		return shim.Error(string(respond))
	}

	//接收参数
	preOrg := args[0]
	subName := args[1]
	newOrg := args[2]
	eventFilter := args[3]


	//参数合法性 —— 原用户是否存在(key)
	prekey := UsrPreString + preOrg + subName
	if !IsKeyExist(stub, prekey) {
		respond := RespondConstruct(KeyExisitedError, "The Subject Hasn't Been Registered Yet", nil)
		return shim.Error(string(respond))
	}

	//参数合法性 —— 新Key是否已经存在
	newKey := UsrPreString + newOrg + subName
	if IsKeyExist(stub, newKey) {
		respond := RespondConstruct(KeyExisitedError, "The New Key Is Existed", nil)
		return shim.Error(string(respond))
	}

	//从账本获取用户数据(可合法存入则一定可合法取出 此处无需进行错误处理)
	usrJson, _ := stub.GetState(prekey)
	usr := new(Usr)
	_ = json.Unmarshal(usrJson, usr)

	//修改用户数据并采用新key存储
	usr.Org = newOrg
	usrJson, _ = json.Marshal(usr)
	_ = stub.PutState(newKey, usrJson)

	//从账本中删除旧key上的用户数据
	err :=stub.DelState(prekey)
	if err != nil {
		respond := RespondConstruct(DelStateError, "Error In DelState of PreKey in Func UsrOrgAlter", nil)
		return shim.Error(string(respond))
	}

	//试试账本数据清除后查询会出错吗--------------
	//if IsKeyExist(stub, prekey) {
	//	usrJson, err = stub.GetState(prekey)
	//	respond := RespondConstruct(KeyExisitedError, "The Pre Key Is Still Existed", err.Error())
	//	return shim.Error(string(respond))
	//}
	//usrJson, err = stub.GetState(prekey)
	//respond := RespondConstruct(KeyExisitedError, err.Error(), nil)
	//return shim.Success(respond)

	//注册日志
	txId := stub.GetTxID()
	timeStamp,_ := GetTimeAsTemplate(stub)
	err = LogRegister(stub, UsrOrgAlterAction, "1", txId, timeStamp)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	UpdateLedgerConst(stub, LogKey)

	//注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//返回修改后的用户数据
	respond := RespondConstruct(Success, SuccessMsg, usr)
	return shim.Success(respond)


}

//基于文件名实现对数据文件的富查询 参数={fileName}
func QueryByFileName(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//检查参数个数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "Error Parameters Num of Func QueryByFileName", nil)
		return shim.Error(string(respond))
	}

	//获取参数
	fileName := args[0]

	//构建富查询字符串
	richStr := fmt.Sprintf(`{"selector":{"data_file_name":"%s"}}`, fileName)

	//查询
	rstIterator, err := stub.GetQueryResult(richStr)
	if err != nil {
		respond := RespondConstruct(RichQueryError, "Error In GetQueryResult Of Func QueryByFileName", err.Error())
		return shim.Error(string(respond))
	}
	defer rstIterator.Close()

	//迭代获取查询结果
	var files []DataFile
	file := new(DataFile)

	for rstIterator.HasNext() {
		fileJson, _ := rstIterator.Next()
		err = json.Unmarshal(fileJson.Value, file)

		//错误处理
		if err != nil {
			respond := RespondConstruct(MarshalError, "Error In UnMarshalling Queried Results Of Func QueryByFileName", err.Error())
			return shim.Error(string(respond))
		}

		files = append(files, *file)
	}

	//构造返回结果
	respond := RespondConstruct(Success, SuccessMsg, files)
	return shim.Success(respond)

}


//通过创建者查询文件 参数={“Creator”, "eventFilter"}
func QueryByCreator(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//检查参数个数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "Error Parameters Num of Func QueryByCreator", nil)
		return shim.Error(string(respond))
	}

	//获取参数
	creator := args[0]


	//构建富查询字符串
	richStr := fmt.Sprintf(`{"selector":{"creator":"%s", "$or":[{"data_file_type":"OnChain"},{"data_file_type":"OffChain"}]}}`, creator)

	//查询
	rstIterator, err := stub.GetQueryResult(richStr)
	if err != nil {
		respond := RespondConstruct(RichQueryError, "Error In GetQueryResult Of Func QueryByCreator", err.Error())
		return shim.Error(string(respond))
	}
	defer rstIterator.Close()

	//迭代获取查询结果
	var files []DataFile
	file := new(DataFile)

	for rstIterator.HasNext() {
		item, _ := rstIterator.Next()
		fileJson := item.Value
		err = json.Unmarshal(fileJson, file)

		//错误处理
		if err != nil {
			respond := RespondConstruct(MarshalError, "Error In UnMarshalling Queried Results Of Func QueryByCreator", err.Error())
			return shim.Error(string(respond))
		}

		files = append(files, *file)
	}


	//构造返回结果
	respond := RespondConstruct(Success, SuccessMsg, files)
	return shim.Success(respond)

}

//日志查询
func GetLog(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//判断参数个数
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "Error Num of parameters in func GetLog", nil)
		return shim.Error(string(respond))
	}

	//构造查询起始key和查询终止key
	start_key := LogPreString + strconv.FormatUint(0, 10)
	endId, _ := GetLedgerConst(stub, LogKey)
	end_key := LogPreString + strconv.FormatUint(endId, 10)

	//获取日志
	var logs []LogInfo
	log := new(LogInfo)

	rstIterator, _ := stub.GetStateByRange(start_key, end_key)

	for rstIterator.HasNext() {
		item, _ := rstIterator.Next()
		logJson := item.Value

		err := json.Unmarshal(logJson, log)
		//错误处理
		if err != nil {
			respond := RespondConstruct(MarshalError, "Error In UnMarshalling Queried Results Of Func QueryByCreator", err.Error())
			return shim.Error(string(respond))
		}

		logs = append(logs, *log)
	}

	//构造响应
	respond := RespondConstruct(Success, SuccessMsg, logs)
	return shim.Success(respond)
}

//目录名更新
func DirNameAlter(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//参数验证
	if len(args) != 4 {
		respond := RespondConstruct(ArgsNumberError, "Error in Params Num of Func DirNameAlter", nil)
		return  shim.Error(string(respond))
	}

	//获取参数list
	parentDir := args[0]
	preName := args[1]
	newName := args[2]
	eventFilter := args[3]

	//验证parentDir是否存在 以及 parentDir下是否有该目录
	if !IsKeyExist(stub, parentDir) {
		respond := RespondConstruct(KeyExisitedError, "No Such ParenDir of Func DirNameAlter", nil)
		return shim.Error(string(respond))
	}

	parentJson, _ := stub.GetState(parentDir)
	parent := new(Dir)
	err := json.Unmarshal(parentJson, parent)
	if err != nil {
		respond := RespondConstruct(UnMarshalError, "The Src Is Not Dir", err.Error())
		return  shim.Error(string(respond))
	}

	flag := false
	var key int
	content := parent.Content
	for index, child := range content {
		if child.SrcName == preName && child.SrcType == "Dir"{
			flag = true
			key = index
			break
		}
	}

	if !flag {
		respond := RespondConstruct(ChildExistedError, fmt.Sprintf("No Child Dir %s in ParentDir %s", preName, parentDir), nil)
		return shim.Error(string(respond))
	}

	//更新DirName分三部分：1.修改其父目录content中的内容 2.修改其本身fileName及absolutePath 3.修改其孩子资源的parentDir(递归)
	//1.修改其父目录content中的内容
	newchild := Child{
		SrcName: newName,
		SrcType: "Dir",
	}
	parent.Content[key] = newchild

	//2\3 修改该目录下子资源的内容及该目录资源本身的信息（通过调用ChangeDirName）
	preAbPath := parentDir + "/" + preName
	newAbPath := parentDir + "/" + newName

	newParentPath := parentDir
	err = ChangeDirName(stub, preAbPath, newParentPath, newAbPath, 1)
	if err != nil {
		respond := RespondConstruct(ChangeChildDirError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//更新父目录信息
	parentJson, _ = json.Marshal(parent)
	_ = stub.PutState(parent.AbsolutePath, parentJson)

	//注册日志
	txId := stub.GetTxID()
	timeStamp,_ := GetTimeAsTemplate(stub)
	err = LogRegister(stub, DirNameAlterAction, "1", txId, timeStamp)
	if err != nil {
		respond := RespondConstruct(LogRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}
	UpdateLedgerConst(stub, LogKey)

	//注册链码事件进行监听
	err = stub.SetEvent(eventFilter, []byte{})
	if err != nil {
		respond := RespondConstruct(EventRegisterError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//返回结果
	respond := RespondConstruct(Success, SuccessMsg, nil)
	return shim.Success(respond)

}


//目录展示
func DirView(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//参数个数判断
	if len(args) != 1 {
		respond := RespondConstruct(ArgsNumberError, "Error Params Num of Func DirView", nil)
		return shim.Error(string(respond))
	}

	//初始目录: args[0]
	allRet := new(DirRet)
	OriginalPath := args[0]
	allRet.SrcAbPath = OriginalPath
	allRet.SrcType = "Dir"

	//创建DirView
	err := DirRetConstruct(stub, allRet, OriginalPath)
	if err != nil {
		respond := RespondConstruct(DirViewConstructError, err.Error(), nil)
		return shim.Error(string(respond))
	}

	//返回结果
	respond := RespondConstruct(Success, SuccessMsg, allRet)
	return shim.Success(respond)

}