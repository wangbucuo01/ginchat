package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	userNum       int
	loginInterval time.Duration
	msgInterval   time.Duration
)

func init() {
	flag.IntVar(&userNum, "u", 500, "登录用户数")
	flag.DurationVar(&loginInterval, "l", 5e9, "用户登陆时间间隔")
	flag.DurationVar(&msgInterval, "m", 1*time.Minute, "用户发送消息时间间隔")
}

func main() {
	flag.Parse()
	for i := 0; i < userNum; i++ {
		go UserConnect("user" + strconv.Itoa(i))
		time.Sleep(loginInterval)
	}
	select {}
}

func UserConnect(userId string) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	var dialer *websocket.Dialer
	token := "123"
	conn, _, err := dialer.Dial("ws://127.0.0.1:8081/chat?userId="+userId+"&token="+token, nil)
	if err != nil {
		fmt.Println(userId, "连接建立失败, err:", err)
		return
	} else {
		fmt.Println(userId,"连接建立成功")
	}
	defer conn.Close()
}
