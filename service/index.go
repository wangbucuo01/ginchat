package service

// TODO: 响应信息的结构统一化

import (
	"html/template"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ginchat/models"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} string
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("index.html", "views/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")

	// c.JSON(200, gin.H{
	// 	"message": "welcome!",
	// })
}

func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")
}

func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/createcom.html",
		"views/chat/main.html")
	if err != nil {
		panic(err)
	}
	userId := c.Query("userId")
	token := c.Query("token")
	user := models.UserBasic{}
	uid, _ := strconv.Atoi(userId)
	user.ID = uint(uid)
	user.Identity = token
	ind.Execute(c.Writer, user)
}

func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}