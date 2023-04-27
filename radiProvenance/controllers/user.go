package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"radiProvenance/models"
)

type UserController struct {
	beego.Controller
}


//错误处理界面
func (c *UserController)ErrHandle() {
	c.Data["errmsg"]= c.GetString("errmsg")
	c.Data["prepage"] = c.GetString("prepage")
	c.TplName = "err.html"
}

//注册界面显示
func (c *UserController)RegisterShow() {
	c.TplName = "register.html"
}

//注册处理
func (c *UserController)RegisterHandle() {

	/*接收form表单数据*/
	userName := c.GetString("userName")
	passWord := c.GetString("passWord")

	/*数据合法性判断*/
	//注册用户名和密码不能为空
	if userName == "" || passWord == "" {
		logs.Info("用户名不能为空")
		c.Data["errmsg"] = "用户名和密码不能为空！"
		c.Data["prepage"] = "/register"
		c.TplName = "err.html"
		return
	}

	//判断用户名是否已经存在
	user := &models.User{
		UserName:userName,
	}
	ormer := orm.NewOrm()
	err := ormer.Read(user, "UserName")
	if err == nil {
		logs.Info("用户名已存在:")
		c.Data["errmsg"] = "用户名已存在！"
		c.Data["prepage"] = "/register"
		c.TplName = "register.html"
		return
	}

	/*注册用户 插库*/
	user.UserName = userName
	user.PassWord = passWord

	_, err = ormer.Insert(user)
	if err != nil {
		logs.Error("注册用户失败：", err.Error())
		c.Data["errmsg"] = "注册用户失败"
		c.Data["prepage"] = "/register"
		c.TplName = "err.html"
		return
	}

	/*返回视图*/
	logs.Info("注册成功")
	c.Redirect("/login", 302)

}

//登录界面展示
func (c *UserController)LoginShow() {
	c.TplName = "login.html"
}

//登录处理
func (c *UserController)LoginHandle() {
	/*获取传入参数*/
	userName := c.GetString("userName")
	passWord := c.GetString("passWord")

	/*查表判断*/
	user := &models.User{
		UserName:     userName,
	}
	ormer := orm.NewOrm()
	err := ormer.Read(user, "UserName")
	if err != nil {
		logs.Error("用户名错误", err.Error())
		c.Data["errmsg"] = "用户名错误"
		c.Data["prepage"] = "/login"
		c.TplName = "err.html"
		return

	}
	if passWord != user.PassWord {
		logs.Error("密码错误", err.Error())
		c.Data["errmsg"] = "密码错误"
		c.Data["prepage"] = "/login"
		c.TplName = "err.html"
		return
	}

	/*登陆成功 保存session*/
	logs.Info("登录成功")
	c.SetSession("userName", userName)

	/*返回视图*/
	c.Redirect("/trace/showMetas", 302)
}

/*退出登录*/
func (c *UserController)Logout() {
	/*消除session*/
	c.DelSession("userName")

	/*跳转登录界面*/
	c.Redirect("/login", 302)
}

