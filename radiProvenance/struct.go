package main


//数据集元信息结构体
type MetaData struct {
	DataId string `json:"data_id"`
	DataName string `json:"data_name"`
	Abstract string `json:"abstract"`
	Owner string `json:"owner"`
	Hash string `json:"hash"`
	DataAddr string `json:"data_addr"`  //IPFS地址
	TimeStamp string `json:"time_stamp"`
	DelFlag string `json:"del_flag"`
	Type string `json:"type"`
	TxId string `json:"tx_id"`
}

//日志信息
type LogInfo struct {
	DataId string `json:"data_id"`
	DataName string `json:"data_name"`
	Owner string `json:"owner"`
	Operator string `json:"operator"`
	Operation string `json:"operation"`
	TimeStamp string `json:"time_stamp"`
	TxId string `json:"tx_id"`
	Type string `json:"type"`
}

//返回值结构体
type Respond struct {
	RespondCode		string `json:"respond_code"`
	RespondMsg		string `json:"respond_msg"`
	RespondData		interface{} `json:"respond_data"`
}

//日志返回结构体
type LogRespond struct {
	LogInfos []LogInfo `json:"log_infos"`
}

//MetaData返回结构体
type MetaRespond struct {
	MetaDatas []MetaData `json:"meta_datas"`
}


//LogId 的账本常量
type IdConst struct {
	LogConst		uint64 `json:"log_const"`
}

//响应码常量
const (
	Success = "0000"
	SuccessMsg = "SmartContract Successfully Exec"

	PutStateError = "0001"
	MarshalError = "0002"
	GetConstIDError = "0003"
	KeyExisitedError = "0004"
	RespondConstructError = "0005"
	GetTimeStampError = "0006"
	ParentDirNotExistedError = "0007"
	AddSrcToParentDirError = "0008"
	ArgsNumberError = "0009"
	FuncNameUndefined = "0010"
	EventRegisterError = "0011"
	DelStateError = "0012"
	RichQueryError = "0013"
	LogRegisterError = "0014"
	ChildExistedError = "0015"
	UnMarshalError = "0016"
	ChangeChildDirError = "0017"
	DirViewConstructError = "0018"
	CompositekeyError = "0019"
	GetStateError = "0020"
)

//前缀常量
const (
	MetaPre = "Meta_"
	LogPre = "Log_"
	IdAttr = "Id_"
)

//类型常量
const (
	MetaType = "MetaData"
	LogType = "LogInfo"
)


//LogConstId存储键
const (
	ConstIdKey = "ledgerIdConst"	//在账本上存储Idconst变量的键值对中的键
)