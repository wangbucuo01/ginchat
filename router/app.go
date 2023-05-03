package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ginchat/docs"
	"github.com/ginchat/service"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()
	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 静态资源
	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/**/*")

	// 首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)

	// 用户模块
	r.GET("/user/list", service.GetUserList)
	r.POST("/user/create", service.CreateUser)
	r.POST("/user/delete", service.DeleteUser)
	r.POST("/user/update", service.UpdateUser)
	r.POST("/user/login", service.FindUserByNameAndPassword)

	// 朋友模块
	r.POST("/friend/find", service.FindByID)
	r.POST("/friends/search", service.SearchFriend)
	r.POST("/friend/add", service.AddFriend)

	// 消息模块
	r.GET("/msg/send", service.SendMsg)
	r.GET("/user/msg/send", service.SendUserMsg)
	r.POST("/user/redisMsg", service.RedisMsg)

	//上传文件
	r.POST("/file/upload", service.Upload)

	// 群聊
	r.POST("community/create", service.CreateCommunity)
	r.POST("community/load", service.LoadCommunity)
	r.POST("community/join", service.JoinCommunity)

	return r
}
