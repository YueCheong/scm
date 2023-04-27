package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"io/ioutil"
	"os"
	"path"
	"radiProvenance/ccQueryandInvoke"
	"radiProvenance/models"
	"radiProvenance/utils"
	"strconv"
	"time"
)

type TraceController struct {
	beego.Controller
}


/*从区块链获取全部数据集的原信息并展示*/
func (c *TraceController)ShowMetas() {

	/*获取链上数据集的MetaInfos*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowAll", []string{})
	if err != nil {
		logs.Error("从链上获取数据失败:", err.Error())
		c.Data["errmsg"] = "链上获取数据失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return

	}

	/*反序列化响应字节码*/
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		c.Data["errmsg"] = "链上数据反序列化失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	jsonBytes := respond.RespondData

	metas := new(models.MetaRespond)
	err = json.Unmarshal(jsonBytes, metas)
	if err != nil {
		logs.Error("元信息数组反序列化失败", err.Error())
		c.Data["errmsg"] = "元信息数组反序列化失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	logs.Info("元信息数据反序列化成功")

	//排序
	length := len(metas.MetaDatas)
	var temp int
	var tempvalue int
	for i:=0; i<length; i++ {
		temp = i
		idi, _ := strconv.Atoi(metas.MetaDatas[i].DataId)
		tempvalue = idi
		for j := i + 1; j < length; j++ {
			idj, _ := strconv.Atoi(metas.MetaDatas[j].DataId)
			if tempvalue > idj {
				tempvalue = idj
				temp = j
			}
		}

		if temp != i {
			tempMeta := metas.MetaDatas[i]
			metas.MetaDatas[i] = metas.MetaDatas[temp]
			metas.MetaDatas[temp] = tempMeta
		}

	}
		/*返回视图*/

	c.Data["userName"] = c.GetSession("userName")
	c.Data["metaInfos"] = metas.MetaDatas
	c.TplName = "index.html"
}


//上传数据页面展示
func (c *TraceController)RegisterDataSetShow() {

	//获取session
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	//返回视图
	c.Data["userName"] = userName
	c.TplName = "addDataSet.html"
}

//上传数据集处理方法
func (c *TraceController)RegisterDataSet() {

	/*接收数据*/
	sessions := c.GetSession("userName")
	userName := sessions.(string)
	dName := c.GetString("dataSetName")
	abstract := c.GetString("abstract")
	f,h,err := c.GetFile("uploadname")

	/*参数合法性判断*/
	//是否接收到上传数据集
	if err != nil {
		logs.Error("未接收到数据集:", err.Error())
		c.Data["errmsg"] = "未接收到数据集"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}
	defer f.Close()

	fileName := h.Filename
	logs.Info("上传文件成功：", fileName)
	logs.Info("数据集名：", dName)
	logs.Info("摘要：", abstract)
	logs.Info("上传者：", userName)

	//数据集名和摘要不能为空
	if dName == ""  || abstract == "" {
		logs.Error("数据集名和摘要不能为空")
		c.Data["errmsg"] = "数据集名和摘要不能为空"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	//数据集名是否重复
	ormer := orm.NewOrm()
	dset := new(models.DataSetInfo)
	dset.DataSetName = dName
	err = ormer.Read(dset, "DataSetName")
	if err == nil {
		logs.Error("数据集名称已存在：")
		c.Data["errmsg"] = "数据集名已存在"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	//上传用户是否合法
	user := &models.User{
		UserName:     userName,
	}
	err = ormer.Read(user, "UserName")
	if err != nil {
		logs.Error("上传者未注册系统:", err.Error())
		c.Data["errmsg"] = "非法用户操作"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	/*AES加密*/
	//生成128bitAES秘钥
	aeskey := utils.Krand(16, models.KC_RAND_KIND_ALL)
	logs.Info("AES秘钥:", aeskey)

	//获取数据集文件内容
	fcontent := make([]byte, h.Size)
	_, err = f.Read(fcontent)
	if err != nil {
		logs.Error("获取文件内容失败：", err.Error())
		c.Data["errmsg"] = "获取数据集文件内容失败"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	//AES加密
	rst, err := utils.AesEncrypt(fcontent, aeskey)
	if err != nil {
		logs.Error("AES加密失败：", err.Error())
		c.Data["errmsg"] = "加密文件失败"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	/*计算加密后哈希*/
	dataHash := utils.Md5(rst)

	/*加密文件bytes上传至IPFS分布式文件系统*/
	ipfsHash, err := utils.UploadIPFS(rst)
	if err != nil {
		logs.Error("IPFS上传失败", err.Error())
		c.Data["errmsg"] = "IPFS上传失败"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	logs.Info("IPFS上传成功_Addr:", ipfsHash)

	/*基础信息存库*/
	//基础信息存库
	dataSetInfo := &models.DataSetInfo{
		DataSetName:dName,
		DataFileName: fileName,
		EncryKey:     string(aeskey),
		User:         user,
	}

	_, err = ormer.Insert(dataSetInfo)
	if err != nil {
		logs.Error("数据集基础信息存库失败：", err.Error())
		c.Data["errmsg"] = "数据集信息存sql库失败"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		return
	}

	/*元信息上链*/
	//获取数据集Id
	err = ormer.Read(dataSetInfo, "DataSetName")
	if err != nil {
		logs.Error("数据集信息获取失败", err.Error())
		return
	}

	id := strconv.FormatUint(uint64(dataSetInfo.Id), 10)
	metaInfo := &models.MetaData{
		DataId:    id,
		DataName:  dName,
		Abstract:  abstract,
		Owner:     userName,
		Hash:      dataHash,
		DataAddr:  ipfsHash,
	}

	//与链交互
	metadata1args := []string{metaInfo.DataId, metaInfo.DataName,
		metaInfo.Abstract, metaInfo.Owner,metaInfo.Hash, metaInfo.DataAddr}

	_, err = ccQueryandInvoke.CCExecute(models.ResClient,
		"MetaRegister", "radiTraceCC", "MetaRegister", metadata1args)
	if err != nil {
		logs.Error("元信息注册失败：", err.Error())
		c.Data["errmsg"] = "元信息上链失败"
		c.Data["prepage"] = "/trace/registerDataSet"
		c.TplName = "err.html"
		_, _ = ormer.Delete(&models.User{Id:dataSetInfo.Id})
		return
	}




	/*返回视图*/
	c.Redirect("/trace/showMyMetas", 302)



}

//展示用户自己的数据集列表
func (c *TraceController)ShowMyMetas()  {

	//获取参数
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	//与链码交互获取数据集信息
	/*获取链上数据集的MetaInfos*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowByOwner", []string{userName})
	if err != nil {
		logs.Error("从链上获取数据失败:", err.Error())
		return
	}

	/*反序列化响应字节码*/
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		return
	}
	jsonBytes := respond.RespondData

	metas := new(models.MetaRespond)
	err = json.Unmarshal(jsonBytes, metas)
	if err != nil {
		logs.Error("元信息数组反序列化失败", err.Error())
		return
	}
	logs.Info("元信息数据反序列化成功")

	//排序
	length := len(metas.MetaDatas)
	var temp int
	var tempvalue int
	for i:=0; i<length; i++{
		temp = i
		idi,_ := strconv.Atoi(metas.MetaDatas[i].DataId)
		tempvalue = idi
		for j:=i+1; j<length; j++{
			idj, _ := strconv.Atoi(metas.MetaDatas[j].DataId)
			if tempvalue > idj{
				tempvalue = idj
				temp = j
			}
		}

		if temp != i{
			tempMeta := metas.MetaDatas[i]
			metas.MetaDatas[i] = metas.MetaDatas[temp]
			metas.MetaDatas[temp] = tempMeta
		}

	}

	//返回视图
	c.Data["userName"] = userName
	c.Data["metaInfos"] = metas.MetaDatas
	c.TplName="mydataset.html"
}

//我的数据集详细信息展示
func (c *TraceController)ShowMyDetails() {

	/*获取传入参数*/
	dName := c.GetString("dataName")
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	/*参数合法性验证*/
	ormer := orm.NewOrm()
	dataSetInfo := models.DataSetInfo{
		DataSetName:  dName,
	}

	var metaInfos []models.DataSetInfo
	_, err := ormer.QueryTable("DataSetInfo").RelatedSel("User").Filter("DataSetName", dName).All(&metaInfos)
	//err := ormer.Read(dataSetInfo, "DataSetName")
	if err != nil {
		logs.Error("从sql库获取数据失败", err.Error())
		return
	}

	dataSetInfo = metaInfos[0]
	if userName != dataSetInfo.User.UserName {
		logs.Error("非法访问他人数据:", userName, " != ", dataSetInfo.User.UserName)
		return
	}

	/*与区块链交互 获取数据详情*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowByDataName", []string{dName})
	if err != nil {
		logs.Error("从链上获取数据失败:", err.Error())
		return
	}

	/*反序列化响应字节码*/
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		return
	}
	jsonBytes := respond.RespondData

	metas := new(models.MetaRespond)
	err = json.Unmarshal(jsonBytes, metas)
	if err != nil {
		logs.Error("元信息数组反序列化失败", err.Error())
		return
	}
	logs.Info("元信息数据反序列化成功")

	if len(metas.MetaDatas) != 1 {
		logs.Error("返回数据集个数错误：", len(metas.MetaDatas))
		return
	}

	meta := metas.MetaDatas[0]

	//返回视图
	c.Data["userName"] = userName
	c.Data["dataId"] = meta.DataId
	c.Data["dataName"] = meta.DataName
	c.Data["abstract"] = meta.Abstract
	c.Data["owner"] = meta.Owner
	c.Data["hash"] = meta.Hash
	c.Data["ipfs"] = meta.DataAddr
	c.Data["timestamp"] = meta.TimeStamp
	c.Data["del"] = meta.DelFlag
	c.Data["txid"] = meta.TxId
	c.Data["encrykey"] = dataSetInfo.EncryKey
	c.TplName = "myDetails.html"

}

//以游客身份看数据集详情
func (c *TraceController) ShowAllDetails() {

	/*获取传入参数*/
	dName := c.GetString("dataName")
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	/*合法性验证*/
	if dName == "" {
		logs.Error("查询数据集名不能为空")
		return
	}

	/*与链交互 获取数据*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowByDataName", []string{dName})
	if err != nil {
		logs.Error("从链上获取数据失败:", err.Error())
		return
	}

	/*反序列化响应字节码*/
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		return
	}
	jsonBytes := respond.RespondData

	metas := new(models.MetaRespond)
	err = json.Unmarshal(jsonBytes, metas)
	if err != nil {
		logs.Error("元信息数组反序列化失败", err.Error())
		return
	}
	logs.Info("元信息数据反序列化成功")

	if len(metas.MetaDatas) != 1 {
		logs.Error("返回数据集个数错误：", len(metas.MetaDatas))
		return
	}

	meta := metas.MetaDatas[0]

	/*返回视图*/
	c.Data["userName"] = userName
	c.Data["dataId"] = meta.DataId
	c.Data["dataName"] = meta.DataName
	c.Data["abstract"] = meta.Abstract
	c.Data["owner"] = meta.Owner
	c.Data["hash"] = meta.Hash
	c.Data["ipfs"] = meta.DataAddr
	c.Data["timestamp"] = meta.TimeStamp
	c.Data["del"] = meta.DelFlag
	c.Data["txid"] = meta.TxId
	c.TplName = "allDetails.html"
}

//下载数据集
func (c *TraceController)Download() {
	/*获取参数*/
	aesKey := c.GetString("aesKey")
	dataId, err := c.GetInt("dataId")
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	if err != nil {
		logs.Error("id不是整数", err.Error())
		c.Data["errmsg"] = "数据集Id非法"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	ipfs := c.GetString("ipfs")

	/*判断下载禁止位*/
	id :=strconv.Itoa(dataId)
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowMetaById", []string{id})
	if err != nil {
		logs.Error("获取数据失败", err.Error())
		c.Data["errmsg"] = "获取链上数据集元信息失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	//反序列化响应字节码
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		c.Data["errmsg"] = "链上数据反序列化失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	jsonBytes := respond.RespondData

	metaInfo := new(models.MetaData)
	err = json.Unmarshal(jsonBytes, metaInfo)
	if err != nil {
		logs.Error("数据集元信息反序列化失败", err.Error())
		c.Data["errmsg"] = "数据集元信息反序列化失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	if metaInfo.DelFlag == "1" {
		logs.Error("此文件禁止下载")
		c.Data["errmsg"] = "此文件状态为禁止下载"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	/*判断密钥*/
	ormer := orm.NewOrm()
	var datasetInfos []models.DataSetInfo
	_, err = ormer.QueryTable("DataSetInfo").RelatedSel("User").Filter("Id", dataId).All(&datasetInfos)
	if err != nil {
		logs.Info(err.Error())
		return
	}

	if len(datasetInfos) != 1 {
		logs.Error("数据集Id不唯一")
		c.Data["errmsg"] = "数据集Id不唯一"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	dataSetInfo := datasetInfos[0]


	if dataSetInfo.EncryKey != aesKey {
		logs.Error("AES密钥错误：", dataSetInfo.EncryKey, " != ", aesKey)
		c.Data["errmsg"] = "AES密钥错误"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	/*从ipfs获取数据*/
	content, err := utils.CatIPFS(ipfs)
	if err != nil {
		logs.Error("IPFS下载失败", err.Error())
		c.Data["errmsg"] = "从IPFS获取文件内容失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}


	/*解密返回*/
	decry, err := utils.AesDecrypt(content, []byte(aesKey))
	if err != nil {
		logs.Error("解密失败", err.Error())
		c.Data["errmsg"] = "AES解密失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}
	//保存文件
	err = ioutil.WriteFile("./static/dataSet/" + dataSetInfo.DataFileName, decry, 0666)
	if err != nil {
		logs.Error("写文件失败", err.Error())
		c.Data["errmsg"] = "写文件失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	c.Ctx.Output.Download("./static/dataSet/" + dataSetInfo.DataFileName, dataSetInfo.DataFileName)
	//传给用户后删除本地文件
	err = os.Remove("./static/dataSet/" + dataSetInfo.DataFileName)
	if err != nil {
		logs.Error("删除失败", err.Error())
		c.Data["errmsg"] = "临时文件删除失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

	/*向区块链注册下载信息*/

	dName := dataSetInfo.DataSetName
	owner := dataSetInfo.User.UserName
	operator := userName


	_, err = ccQueryandInvoke.CCExecute(models.ResClient, "download", "radiTraceCC", "DataDownload", []string{id, dName, owner, operator})
	if err != nil {
		logs.Error("调用下载链码失败！", err.Error())
		c.Data["errmsg"] = "调用下载链码失败"
		c.Data["prepage"] = "/trace/showMetas"
		c.TplName = "err.html"
		return
	}

}


//删除位修改
func (c *TraceController)DelFlagConvert() {

	/*从前端获取数据*/
	dataId := c.GetString("dataId")
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	/*参数判断*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowMetaById", []string{dataId})
	if err != nil {
		logs.Error("获取数据失败", err.Error())
		c.Data["errmsg"] = "调用下载链码失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	//反序列化响应字节码
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		c.Data["errmsg"] = "链上数据反序列化失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	jsonBytes := respond.RespondData

	metaInfo := new(models.MetaData)
	err = json.Unmarshal(jsonBytes, metaInfo)
	if err != nil {
		logs.Error("数据集元信息反序列化失败", err.Error())
		c.Data["errmsg"] = "数据集元信息反序列化失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	//判断操作用户是否合法
	if userName != metaInfo.Owner {
		logs.Error("非法用户操作：", userName, " != ", metaInfo.Owner)
		c.Data["errmsg"] = "非法用户操作"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	/*修改数据集DelFlag标志位*/
	_, err = ccQueryandInvoke.CCExecute(models.ResClient, "delManage", "radiTraceCC", "DelData", []string{dataId})
	if err != nil {
		logs.Error("账本DelFlag操作失败：", err.Error())
		c.Data["errmsg"] = "区块链delFlag信息更新失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	c.Redirect("/trace/showMyMetas", 302)
}

//更新数据集界面展示
func (c *TraceController)ShowAlter() {


	/*获取参数*/
	dataId, err := c.GetInt("dataId")
	if err != nil {
		logs.Error("获取ID失败", err.Error())
		return
	}
	id := strconv.Itoa(dataId)
	sessions := c.GetSession("userName")
	userName := sessions.(string)


	/*查询数据集详情*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowMetaById", []string{id})
	if err != nil {
		logs.Error("获取数据失败", err.Error())
		return
	}

	//反序列化响应字节码
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		return
	}

	jsonBytes := respond.RespondData

	metaInfo := new(models.MetaData)
	err = json.Unmarshal(jsonBytes, metaInfo)
	if err != nil {
		logs.Error("数据集元信息反序列化失败", err.Error())
		return
	}


	/*判断用户操作是否合法*/
	if metaInfo.Owner != userName {
		logs.Error("非法用户操作：", metaInfo.Owner, " != ", userName)
		return
	}

	/*返回视图*/
	c.Data["userName"] = userName
	c.Data["dataName"] = metaInfo.DataName
	c.Data["abstract"] = metaInfo.Abstract
	c.Data["dataId"] = metaInfo.DataId
	c.TplName = "alterDataSet.html"
}

//更新数据集post处理
func (c *TraceController)AlterPost() {

	/*获取参数*/
	abstract := c.GetString("abstract")
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	dataId, err :=c.GetInt("dataId")
	if err != nil {
		logs.Error("获取Id失败：", err.Error())
		c.Data["errmsg"] = "获取Id失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}
	id := strconv.Itoa(dataId)


	f,h,err := c.GetFile("uploadname")
	flag := 0  //用户是否上传新数据集判断标志位

	if err != nil {
		logs.Info("用户未上传新数据集")
	}else {
		logs.Info("上传新数据集")
		flag = 1
	}


	/*参数合法性判断*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowMetaById", []string{id})
	if err != nil {
		logs.Error("获取数据失败", err.Error())
		c.Data["errmsg"] = "sql库获取数据失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	//反序列化响应字节码
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		c.Data["errmsg"] = "链上数据反序列化失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	jsonBytes := respond.RespondData

	metaInfo := new(models.MetaData)
	err = json.Unmarshal(jsonBytes, metaInfo)
	if err != nil {
		logs.Error("数据集元信息反序列化失败", err.Error())
		c.Data["errmsg"] = "数据集元信息反序列化失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	//判断用户操作是否合法
	if metaInfo.Owner != userName {
		logs.Error("非法用户操作：", metaInfo.Owner, " != ", userName)
		c.Data["errmsg"] = "非法用户操作"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	//空字符串判断
	if abstract == "" {
		logs.Error("数据集名和摘要不能为空")
		c.Data["errmsg"] = "数据集名和摘要不能为空"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}


	/*修改*/
	//获取sql中数据集信息
	ormer := orm.NewOrm()
	dataSetInfo := new(models.DataSetInfo)
	dataSetInfo.Id = uint(dataId)

	err = ormer.Read(dataSetInfo)
	if err != nil {
		logs.Error("sql中没有Id为", id, "的数据集")
		c.Data["errmsg"] = "sql中没有此Id的数据集信息"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	if flag == 1 { //上传了新的数据集

		//获取加密秘钥
		aeskey := []byte(dataSetInfo.EncryKey)


		//获取数据集文件内容
		fcontent := make([]byte, h.Size)
		_, err = f.Read(fcontent)
		if err != nil {
			logs.Error("获取文件内容失败：", err.Error())
			c.Data["errmsg"] = "获取文件内容失败"
			c.Data["prepage"] = "/trace/showMyMetas"
			c.TplName = "err.html"
			return
		}

		//AES加密
		rst, err := utils.AesEncrypt(fcontent, aeskey)
		if err != nil {
			logs.Error("AES加密失败：", err.Error())
			c.Data["errmsg"] = "AES加密失败"
			c.Data["prepage"] = "/trace/showMyMetas"
			c.TplName = "err.html"
			return
		}

		/*计算加密后哈希*/
		dataHash := utils.Md5(rst)

		/*加密文件bytes上传至IPFS分布式文件系统*/
		ipfsHash, err := utils.UploadIPFS(rst)
		if err != nil {
			logs.Error("IPFS上传失败", err.Error())
			c.Data["errmsg"] = "IPFS上传文件失败"
			c.Data["prepage"] = "/trace/showMyMetas"
			c.TplName = "err.html"
			return
		}
		logs.Info("IPFS上传成功_Addr:", ipfsHash)

		/*修改sql库数据*/
		dataSetInfo.DataFileName = h.Filename
		_, err = ormer.Update(dataSetInfo, "DataFileName")
		if err != nil {
			logs.Error("更新sql数据库失败")
			c.Data["errmsg"] = "更新sql数据库失败"
			c.Data["prepage"] = "/trace/showMyMetas"
			c.TplName = "err.html"
			return
		}

		/*修改metaInfo信息*/
		metaInfo.Hash = dataHash
		metaInfo.DataAddr = ipfsHash
	}

	/*更新链上数据*/
	metaInfo.Abstract = abstract
	_, err = ccQueryandInvoke.CCExecute(models.ResClient, "dataAlter", "radiTraceCC",
		"MetaAlter", []string{metaInfo.DataId, metaInfo.DataName,
			metaInfo.Abstract, metaInfo.Owner, metaInfo.Hash, metaInfo.DataAddr})
	if err != nil {
		logs.Error("链上数据更新失败")
		c.Data["errmsg"] = "链上数据更新失败"
		c.Data["prepage"] = "/trace/showMyMetas"
		c.TplName = "err.html"
		return
	}

	logs.Info("更新成功")

	/*返回视图*/
	c.Redirect("/trace/showMyMetas", 302)
}

func (c *TraceController)ShowLogs() {

	/*获取参数*/
	sessions := c.GetSession("userName")
	userName := sessions.(string)
	dataId := c.GetString("dataId")

	/*参数判断*/
	ormer := orm.NewOrm()
	var dataSetInfos []models.DataSetInfo
	_, err := ormer.QueryTable("DataSetInfo").RelatedSel("User").Filter("Id", dataId).All(&dataSetInfos)
	if err != nil {
		logs.Error("没有Id为", dataId, "的数据")
		return
	}
	//用户操作是否合法
	dataSetInfo := dataSetInfos[0]
	if userName != dataSetInfo.User.UserName {
		logs.Error("非法用户操作：", userName, " != ", dataSetInfo.User.UserName)
		return
	}

	/*查询日志*/
	responds, err := ccQueryandInvoke.QueryCC(models.ResClient, "radiTraceCC", "ShowLogsById", []string{dataId})
	if err != nil {
		logs.Error("获取数据失败", err.Error())
		return
	}

	//反序列化响应字节码
	respond := new(models.Respond)
	err = json.Unmarshal(responds.Payload, respond)
	if err != nil {
		logs.Error("链上数据反序列化失败：", err.Error())
		return
	}

	jsonBytes := respond.RespondData

	logRespond := new(models.LogRespond)
	err = json.Unmarshal(jsonBytes, logRespond)
	if err != nil {
		logs.Error("日志元信息反序列化失败", err.Error())
		return
	}


	//排序
	length := len(logRespond.LogInfos)
	var temp int
	var tempvalue int64
	for i:=0; i<length; i++ {
		temp = i
		t, _ := time.ParseInLocation("2006-01-02", logRespond.LogInfos[i].TimeStamp, time.Local)
		idi := t.Unix()
		logs.Info("idi", idi)
		tempvalue = idi
		for j := i + 1; j < length; j++ {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05", logRespond.LogInfos[j].TimeStamp, time.Local)
			idj := t.Unix()
			if tempvalue < idj {
				tempvalue = idj
				temp = j
			}
		}

		if temp != i {
			tempMeta := logRespond.LogInfos[i]
			logRespond.LogInfos[i] = logRespond.LogInfos[temp]
			logRespond.LogInfos[temp] = tempMeta
		}

	}

	c.Data["logs"] = logRespond.LogInfos
	c.Data["userName"] = userName
	c.TplName = "showLogs.html"

}


func (c *TraceController) ShowWaterMarkAdd() {
	/*返回视图*/
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	c.Data["userName"] = userName
	c.TplName = "waterMark.html"

}

func (c *TraceController) WaterMarkAddPost() {

	/*获取数据*/
	waterMarkStr := c.GetString("waterMark")
	f, h, err := c.GetFile("uploadname")

	/*判断参数*/
	//图片接收成功判断
	if err != nil {
		logs.Error("未接收到图片文件")
		c.Data["errmsg"] = "未接收到图片文件"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}
	defer f.Close()

	//水印不为空判断
	if waterMarkStr == "" {
		logs.Error("水印字符串不能为空")
		c.Data["errmsg"] = "水印字符串不能为空"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	//判断图片类型 仅支持.jpg、jpeg和png格式的图片
	ext := path.Ext(h.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		logs.Error("图片格式不支持", ext)
		c.Data["errmsg"] = "仅支持.jpg .jpeg .png格式的图片"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}


	/*保存图片*/
	err = c.SaveToFile("uploadname", "./static/dataSet/" + h.Filename)
	if err != nil {
		logs.Error("写文件失败", err.Error())
		c.Data["errmsg"] = "写文件失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	/*对图片添加水印并保存*/
	waterMark := []byte(waterMarkStr)
	err = utils.WaterMarking("./static/dataSet/"+h.Filename, waterMark, "./static/dataSet/waterMark.png" )
	if err != nil {
		logs.Error("添加水印字符串失败", err.Error())
		c.Data["errmsg"] = "添加水印字符串失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	/*向用户传输该水印图片*/
	c.Ctx.Output.Download("./static/dataSet/waterMark.png" , "waterMark.png")

	//传给用户后删除本地文件
	err = os.Remove("./static/dataSet/" + h.Filename)
	if err != nil {
		logs.Error("删除图片失败", err.Error())
		c.Data["errmsg"] = "临时文件删除失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	err = os.Remove("./static/dataSet/waterMark.png")
	if err != nil {
		logs.Error("删除水印文件失败", err.Error())
		c.Data["errmsg"] = "临时水印文件删除失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

}

func (c *TraceController)WaterMarkDetect() {

	/*获取参数*/
	sessions := c.GetSession("userName")
	userName := sessions.(string)

	f, h, err := c.GetFile("uploadname")

	/*判断参数*/
	//图片接收成功判断
	if err != nil {
		logs.Error("未接收到图片文件")
		c.Data["errmsg"] = "未接收到图片文件"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}
	defer f.Close()

	/*保存图片*/
	err = c.SaveToFile("uploadname", "./static/dataSet/" + h.Filename)
	if err != nil {
		logs.Error("写文件失败", err.Error())
		c.Data["errmsg"] = "写文件失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	/*检测水印*/
	waterMarkbytes, err := utils.ReadWaterMark("./static/dataSet/" + h.Filename)
	if err != nil {
		logs.Error("检测水印出错", err.Error())
		c.Data["errmsg"] = "检测水印出错"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
	}

	/*删除临时文件*/
	err = os.Remove("./static/dataSet/" + h.Filename)
	if err != nil {
		logs.Error("删除临时文件失败", err.Error())
		c.Data["errmsg"] = "临时文件删除失败"
		c.Data["prepage"] = "/trace/waterMarkadd"
		c.TplName = "err.html"
		return
	}

	waterMark := string(waterMarkbytes)

	/*返回视图*/
	c.Data["waterMark"] = waterMark
	c.Data["userName"] = userName

	c.TplName = "waterMark.html"

}

