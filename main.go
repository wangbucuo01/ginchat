package main

import (
	_ "net/http/pprof"
	"time"

	"github.com/ginchat/models"
	"github.com/ginchat/router"
	"github.com/ginchat/utils"
	"github.com/spf13/viper"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	InitTimer()

	r := router.Router()
	r.Run(":8081")
}

func InitTimer() {
	utils.Timer(time.Duration(viper.GetInt("time.DelayHeartbeat"))*time.Second, time.Duration(viper.GetInt("timeout.HeartbeadHz"))*time.Second, models.CleanConnection, "")
}
