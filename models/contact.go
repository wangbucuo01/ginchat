package models

import (
	"github.com/ginchat/utils"
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	OwnerId  uint // 谁的关系
	TargetId uint // 和谁的关系 也可对应群id
	Type     int  // 类型 1 好友 2 群组
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.TargetId)
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

func AddFriend(userId uint, targetName string) (int, string) {
	if targetName != "" {
		targetUser := FindUserByName(targetName)
		if targetUser.Salt != "" {
			if targetUser.ID == userId {
				return -1, "不能添加自己为好友"
			}
			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, targetUser.ID).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}
			// 开启一个事务，保证加好友过程的原子性
			tx := utils.DB.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetUser.ID
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			// 添加好友是一个双向关系
			contact1 := Contact{}
			contact1.OwnerId = targetUser.ID
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加好友成功"
		}
		return -1, "没有找到此用户"
	}
	return -1, "好友昵称不能为空"
}
