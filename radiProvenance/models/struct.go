package models

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
	RespondData		[]byte `json:"respond_data"`
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

//用户数据表
type User struct {
	Id uint `orm:"pk;auto"`
	UserName string `orm:"size(50)"`
	PassWord string `orm:"size(50)"`
	DataSetInfos []*DataSetInfo `orm:"reverse(many)"`
}


//DataSet及加密秘钥
type DataSetInfo struct {
	Id uint `orm:"pk;auto"`
	DataSetName string `orm:"size(50)"`
	DataFileName string `orm:"size(50)"`
	EncryKey string `orm:"size(50);null"`
	User *User `orm:"rel(fk)"`
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

//AES秘钥生成控制键
const (
	KC_RAND_KIND_NUM   = 0  // 纯数字
	KC_RAND_KIND_LOWER = 1  // 小写字母
	KC_RAND_KIND_UPPER = 2  // 大写字母
	KC_RAND_KIND_ALL   = 3  // 数字、大小写字母
)