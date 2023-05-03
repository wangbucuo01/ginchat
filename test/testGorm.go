package main

// 通过test创建数据库表，并新建数据

import (
	"github.com/ginchat/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:qhdwsx130324@tcp(127.0.0.1:3306)/ginchat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 如果没有这个表，自动创建
	db.AutoMigrate(&models.UserBasic{})
	db.AutoMigrate(&models.GroupBasic{})
	db.AutoMigrate(&models.Contact{})
	db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.Community{})

	// create
	// user := &models.UserBasic{}
	// user.Name = "王不错"
	// db.Create(user)

	// read
	// fmt.Println(db.First(user, 1)) // 根据整型主键查找
	// db.First(user, "code=?", "D42") // 查找code字段值为D42的记录

	// update
	// db.Model(&user).Update("PassWord", "1234")

}
