package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	// 读取配置文件: 通过viper
	// 读取配置文件名称
	viper.SetConfigName("app")
	// 读取配置文件路径
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("read config error: ", err)
	}
	fmt.Println("config app inited.")
}

func InitMySQL() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢SQL阀值
			LogLevel:      logger.Info, // 级别
			Colorful:      true,        // 彩色
		},
	)
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("config mysql inited.")
}

func InitRedis() {
	Red = redis.NewClient(
		&redis.Options{
			Addr:         viper.GetString("redis.addr"),
			Password:     viper.GetString("redis.password"),
			DB:           viper.GetInt("redis.DB"),
			PoolSize:     viper.GetInt("redis.poolSize"),
			MinIdleConns: viper.GetInt("redis.minIdleConn"),
		},
	)
	fmt.Println("inited redis.")
}

const (
	PublishKey = "websocket"
)

// publish发布消息到redis
func Publish(ctx context.Context, channel string, msg string) error {
	err := Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println("publish to redis err: ", err)
		return err
	}
	return err
}

// subscribe订阅redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println("subcribe to redis err: ", err)
		return "", err
	}
	fmt.Println("subcribe:", msg)
	return msg.Payload, err
}
