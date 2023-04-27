package main

type Dir struct {
	AbsolutePath string `json:"absolute_path"`	//primary key
	ParentDir    string `json:"parent_dir"`		//foreign key
	DirName      string `json:"dir_name"`
	DirId		 string `json:"dir_id"`			//key
	Type         string `json:"type"`
	Size         int    `json:"size"`			//the file number in this Dir
	Content      []Child `json:"content"`        //the file's absolutePath in this Dir
	Creator      string `json:"creator"`        //the creator of the Dir
	ModifyTime   string	`json:"modify_time"`    
	Remark       string `json:"remark"`			
	Attributes   string `json:"attributes"`
	TxId		 string `json:"tx_id"`
}

type DataFile struct {
	AbsolutePath	string `json:"absolute_path"`
	ParentDir		string `json:"parent_dir"`
	DataFileName	string `json:"data_file_name"`
	DataFileId		string `json:"data_file_id"`
	DataFileType    string `json:"data_file_type"`
	Size			string `json:"size"`
	Content			string `json:"content"`
	checkSum		string `json:"chenck_sum"`
	Creator			string `json:"creator"`
	ModifyTime		string `json:"modify_time"`
	Remark          string `json:"remark"`
	Attributes      string `json:"attributes"`
	TxId            string `json:"tx_id"`
}

type Usr struct {
	Org 		string `json:"org"`
	SubjectName string `json:"subject_name"`
	SubjectId   string `json:"subject_id"`
	Pubkey      string `json:"pubkey"`
	Issuer      string `json:"issuer"`
	Expire      string `json:"expire"`
	Signature   string `json:"signature"`
	Attributes  string `json:"attributes"`
	TxId		string `json:"tx_id"`
}

type Child struct {
	SrcName string `json:"src_name"`
	SrcType string `json:"src_type"`
}

type DirRet struct {
	SrcAbPath string `json:"src_ab_path"`
	SrcType   string `json:"src_type"`
	Childs    []DirRet `json:"childs"`
}

type IdConst struct {
	UserIdConst 	uint64 `json:"user_id_const"`
	DirIdConst 		uint64 `json:"dir_id_const"`
	DataFileConst 	uint64 `json:"data_file_const"`
	LogConst		uint64 `json:"log_const"`	
}

type Respond struct {
	RespondCode		string `json:"respond_code"`
	RespondMsg		string `json:"respond_msg"`
	RespondData		interface{} `json:"respond_data"`
}

type LogInfo struct {
	LogAction string `json:"log_action"`
	Decision  string `json:"decision"`
	Txid      string `json:"txid"`
	TimeStamp string `json:"time_stamp"`
}


//资源存储键Key的前缀定义
const (
	UsrPreString = "USR_"
	DirPreString = "Dir_"
	DataFilePreString = "DataFile_"
	LogPreString = "Log_"
)


//三种资源的Id存储键和选择控制符
const (
	ConstIdKey = "ledgerIdConst"	//在账本上存储Idconst变量的键值对中的键
	UserKey = 0						//GetLedgerConst函数使用，选择usrId返回
	DirKey = 1						//GetLedgerConst函数使用，选择DirId返回
	DataFileKey = 2					//GetLedgerConst函数使用，选择DataFileId返回
	LogKey = 3						
)

//资源类型
const (
	DirType = "Dir"
	OnChainType = "OnChain"
	OffChainType = "OffChain"
)


//响应码定义
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
)

const (
	DirRegisterAction string = "DirRegister"
	DataFileRegisterAction string = "DataFileRegister"
	UserRegisterAction string = "UsrRegister"
	UsrOrgAlterAction string = "UsrOrgAlter"
	DirNameAlterAction string = "DirNameAlter"
)


