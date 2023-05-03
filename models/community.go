package models

import (
	"fmt"

	"github.com/ginchat/utils"
	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	Name    string
	OwnerId uint
	Img     string
	Desc    string
}

func FindCommunityByName(name string) Community {
	com := Community{}
	utils.DB.Where("name = ?", name).First(&com)
	return com
}

func CreateCommunity(community Community) (int, string) {
	tx := utils.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if len(community.Name) == 0 {
		return -1, "群名不能为空"
	}
	data := FindCommunityByName(community.Name)
	if data.Name != "" {
		return -1, "群名已存在"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		return -1, "建群失败"
	}
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2 // 群关系
	if err := utils.DB.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "添加群关系失败"
	}
	tx.Commit()
	return 0, "创建群成功"
}

func LoadCommunity(ownerId uint) ([]*Community, string) {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("owner_id = ? and type = 2", ownerId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.TargetId)
	}
	data := make([]*Community, 10)
	utils.DB.Where("id in ?", objIds).Find(&data)
	return data, "查询成功"
}

// 通过群名加群
func JoinCommunity(userId uint, comName string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2
	community := Community{}
	utils.DB.Where("name = ?", comName).Find(&community)
	if community.ID == 0 {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type=2", userId, community.ID).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}

func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.OwnerId)
	}
	return objIds
}
