package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/ginchat/models"
	"github.com/ginchat/utils"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} string
// @Router /user/list [get]
func GetUserList(c *gin.Context) {
	data := models.GetUserList()
	// TODO:分页
	c.JSON(200, gin.H{
		"code":    0,
		"message": "查询用户列表成功",
		"data":    data,
	})
}

// FindUserByNameAndPassword
// @Summary 登录
// @Tags 用户模块
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Success 200 {string} string
// @Router /user/login [post]
func FindUserByNameAndPassword(c *gin.Context) {
	data := models.UserBasic{}
	// 获取参数
	name := c.Request.FormValue("name")
	passwd := c.Request.FormValue("password")
	// 先根据name判断用户是否存在
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "该用户不存在",
			"data":    data,
		})
		return
	}
	// 根据密码判断输入是否正确
	flag := utils.ValidPassword(passwd, user.Salt, user.Password)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "密码不正确",
			"data":    data,
		})
		return
	}
	// 正确的话去查询数据库
	// 先处理密码，因为传入的是明文密码，数据库存的是密文，需要加密处理一下
	passwd = utils.MakePassword(passwd, user.Salt) // 注意这块在创建用户时，给用户加了一个盐的属性用于加密密码
	data = models.FindUserByNameAndPassword(name, passwd)
	c.JSON(200, gin.H{
		"code":    0, //0：成功  -1：失败
		"message": "登录成功",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Param repassword formData string false "确认密码"
// @Success 200 {string} string
// @Router /user/create [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")

	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("repassword")
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户名或密码不能为空!",
			"data":    user,
		})
		return
	}
	if password != repassword {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "两次密码不一致!",
			"data":    user,
		})
		return
	}

	data := models.FindUserByName(user.Name)
	if data.Name != "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户名已注册",
			"data":    user,
		})
		return
	}

	// user.Password = password
	salt := fmt.Sprintf("%06d", rand.Int31())
	user.Salt = salt
	user.Password = utils.MakePassword(password, salt)
	user.LoginTime = time.Now()
	user.LogOutTime = time.Now()
	user.HeartbeatTime = time.Now()
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "新增用户成功!",
		"data":    user,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @Param id query string false "id"
// @Success 200 {string} string
// @Router /user/delete [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "删除用户成功!",
		"data":    user,
	})
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @Param id formData string false "id"
// @Param name formData string false "name"
// @Param password formData string false "password"
// @Param phone formData string false "phone"
// @Param email formData string false "email"
// @Success 200 {string} string
// @Router /user/update [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "修改参数不匹配!",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0,
			"message": "修改用户成功!",
			"data":    user,
		})
	}
}

// 防止跨域站点的伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	// 升级为websocket协议
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("ws upgrade err:", err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)

	MsgHandler(c, ws)
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := models.RedisMsg(uint(userIdA), uint(userIdB), uint(start), uint(end), isRev)
	utils.RespOKList(c.Writer, "ok", res)
}

func MsgHandler(c *gin.Context, ws *websocket.Conn) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println("subscribe err:", err)
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

// SearchFriend
// @Summary 查找所有朋友
// @Tags 朋友模块
// @Param userId formData string false "userId"
// @Success 200 {string} string
// @Router /friends/search [post]
func SearchFriend(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(uid))
	// c.JSON(200, gin.H{
	// 	"code":    0,
	// 	"message": "查询用户列表成功",
	// 	"data":    user,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}

// FindByID
// @Summary 查找用户
// @Tags 朋友模块
// @Param userId formData string false "userId"
// @Success 200 {string} string
// @Router /friend/find [post]
func FindByID(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Request.FormValue("userId"))
	userId := uint(uid)
	user := models.FindByID(userId)
	utils.RespOK(c.Writer, user, "查找成功!")
}

func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	code, msg := models.AddFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// CreateCommunity
// @Summary 创建群聊
// @Tags 群聊模块
// @Param ownerId formData string false "ownerId"
// @Param name formData string false "name"
// @Success 200 {string} string
// @Router /community/create [post]
func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	icon := c.Request.FormValue("icon")
	desc := c.Request.FormValue("desc")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Img = icon
	community.Desc = desc
	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// LoadCommunity
// @Summary 加载群聊
// @Tags 群聊模块
// @Param ownerId formData string false "ownerId"
// @Success 200 {string} string
// @Router /community/load [post]
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := models.LoadCommunity(uint(ownerId))
	if len(data) != 0 {
		utils.RespOKList(c.Writer, data, len(data))
	} else {
		utils.RespFail(c.Writer, msg)
	}

}

// JoinCommunity
// @Summary 加入群聊
// @Tags 群聊模块
// @Param userId formData string false "userId"
// @Param comName formData string false "comName"
// @Success 200 {string} string
// @Router /community/join [post]
func JoinCommunity(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comName := c.Request.FormValue("comName")
	data, msg := models.JoinCommunity(uint(userId), comName)
	if data == 0 {
		utils.RespOK(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
