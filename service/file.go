package service

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/ginchat/utils"
	"github.com/spf13/viper"
)

// 把文件上传到本地
func Upload(c *gin.Context) {
	// UploadLocal(c) // 上传到本地
	// 上传到阿里云OSS
	UploadOSS(c)
}

func UploadLocal(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	suffix := ".png"
	ofilName := head.Filename
	temp := strings.Split(ofilName, ".")
	if len(temp) > 1 {
		suffix = "." + temp[len(temp)-1]
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	utils.RespOK(w, url, "发送图片成功!")
}

func UploadOSS(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	suffix := ".png"
	ofilName := head.Filename
	temp := strings.Split(ofilName, ".")
	if len(temp) > 1 {
		suffix = "." + temp[len(temp)-1]
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	client, err := oss.New(viper.GetString("oss.EndPoint"), viper.GetString("oss.AccessKeyId"), viper.GetString("oss.AccessKeySecret"))
	if err != nil {
		fmt.Println("oss new failed:", err)
		os.Exit(-1)
	}
	// 填写存储空间名称，例如examplebucket。
	bucket, err := client.Bucket(viper.GetString("oss.Bucket"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	err = bucket.PutObject(fileName, srcFile)
	if err != nil {
		fmt.Println("Error:", err)
		utils.RespFail(w, err.Error())
		os.Exit(-1)
	}

	url := "http://" + viper.GetString("oss.Bucket") + "." + viper.GetString("oss.EndPoint") + "/" + fileName
	utils.RespOK(w, url, "发送图片成功")
}
