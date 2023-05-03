package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ginchat/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"

	filter "github.com/antlinker/go-dirtyfilter"
	"github.com/antlinker/go-dirtyfilter/store"
)

type Message struct {
	gorm.Model
	UserId     uint   // 发送者id
	TargetId   uint   // 接受者id
	Type       int    // 发送类型 1 私聊 2 群聊 3 心跳
	Media      int    // 消息类型 1 文字 2 表情包 3 语音 4 图片/表情包 
	Content    string // 消息内容
	CreateTime uint64 // 创建时间
	ReadTime   uint64 // 读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int // 其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

// 一次连接节点
type Node struct {
	Conn          *websocket.Conn //连接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

// 映射关系:用户和他建立的连接的映射
var clientMap map[uint]*Node = make(map[uint]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

func Chat(writer http.ResponseWriter, request *http.Request) {
	// 获取参数，并校验token
	query := request.URL.Query()
	userId := query.Get("userId")
	userid, _ := strconv.Atoi(userId)
	useridu := uint(userid)
	// token := query.Get("token")
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// context := query.Get("context")
	isvalida := true // Todo:checkToken()
	// 将HTTP请求升级为Websocket协议，同时防止跨域站点的伪造请求
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println("连接建立失败:", err)
		return
	}

	// 获取连接
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(),
		HeartbeatTime: currentTime,
		LoginTime:     currentTime,
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	// 用户关系

	// userID
	rwLocker.Lock()
	clientMap[useridu] = node
	rwLocker.Unlock()

	// 完成发送逻辑
	go sendProc(node)
	// 完成接收逻辑
	go recvProc(node)
	// 进入系统的第一条通知
	//sendMsg(useridu, []byte("欢迎进入聊天室"))

	SetUserOnlineInfo("online_"+userId, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
}

func SetUserOnlineInfo(key string, val []byte, timeTTL time.Duration) {
	ctx := context.Background()
	utils.Red.Set(ctx, key, val, timeTTL)
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			// 写入
			fmt.Println("[ws]发送消息:", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			fmt.Println("write data:", string(data))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		// 读数据
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
			return
		}
		// 心跳检测
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			// 更新心跳时间
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			// 将数据进行广播
			broadMsg(data)
			fmt.Println("[ws]接收消息 ", string(data))
		}

	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine...")
}

// 完成udp数据发送协程
func udpSendProc() {
	// 协议，源Ip(nil指本地地址),目的Ip(本地网关)
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: 3000,
	})
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udp send msg:", string(data))
			_, err := conn.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 接收
func udpRecvProc() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		var buf [512]byte
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	// 私信
	case 1:
		fmt.Println("udp recv msg:", string(data))
		sendMsg(msg.TargetId, data)
	// 群发
	case 2:
		sendGroupMsg(msg.TargetId, data)
		// 广播
		// case 3:
		// 	sendAllMsg()
		// case 4:
	}
}

func sendMsg(targetId uint, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[targetId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)

	// 敏感词过滤
	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: []string{"傻子", "坏蛋", "傻缺", "傻屌", "傻大个"},
	})
	if err != nil {
		panic(err)
	}
	filterManage := filter.NewDirtyManager(memStore)
	result, err := filterManage.Filter().Replace(jsonMsg.Content, '*')
	if err != nil {
		panic(err)
	}
	jsonMsg.Content = result

	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(targetId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())

	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	fmt.Println("key: online_"+userIdStr+" value: ", r)
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			// 管道实现消息的接收（消息写入队列）
			node.DataQueue <- msg
		}
	}
	var key string
	if targetId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}

	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	fmt.Println("key:", key, " value(res): ", res)
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	// 使用redis存储聊天记录
	// 使用zset存储聊天记录
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result()
	fmt.Println("key: ", key, " msg: ", string(msg), "score: ", score)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println("ress: ", ress)
}

func sendGroupMsg(targetId uint, msg []byte) {
	fmt.Println("开始群发消息")
	userIds := SearchUserByGroupId(uint(targetId))
	fmt.Println(userIds)
	for i := 0; i < len(userIds); i++ {
		//排除给自己的
		// if targetId != userIds[i] {
		// 	sendMsg(userIds[i], msg)
		// }
		sendMsg(userIds[i], msg)
	}
}

//需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// 聊天记录的回调
// 前端：在点击联系人时，前端调用service.RedisMsg读写历史聊天记录
// 后端：
func RedisMsg(userIdA uint, userIdB uint, start uint, end uint, isRev bool) []string {
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	var rels []string
	var err error
	// 正序输出
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, int64(start), int64(end)).Result()
	} else {
		// 逆序输出
		rels, err = utils.Red.ZRevRange(ctx, key, int64(start), int64(end)).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	return rels
}

//更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 判断心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		timeout = true
	}
	return
}

// 清理超时连接
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix())
	for i := range clientMap {
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时..... 关闭连接：", node)
			node.Conn.Close()
		}
	}
	return result
}
