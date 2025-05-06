package services

import (
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"gorm.io/gorm"
)

// MembershipService 会员服务接口
type MembershipService interface {
	CheckUsageLimit(userID uint) (bool, error)
	IncrementUsage(userID uint) error
	GetMembershipInfo(userID uint) (*models.Membership, error)
}

// MembershipServiceImpl 会员服务实现
type MembershipServiceImpl struct {
	DB *gorm.DB
}

// NewMembershipService 创建会员服务
func NewMembershipService(db *gorm.DB) MembershipService {
	return &MembershipServiceImpl{DB: db}
}

// CheckUsageLimit 检查用户是否超出使用限制
func (s *MembershipServiceImpl) CheckUsageLimit(userID uint) (bool, error) {
	var membership models.Membership
	if err := s.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		return false, err
	}

	// 无限制
	if membership.DailyLimit < 0 {
		return true, nil
	}

	today := time.Now().Format("2006-01-02")
	var usage models.UserUsage
	if err := s.DB.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有今日记录，肯定没超限
			return true, nil
		}
		return false, err
	}

	// 检查是否超限
	return usage.UsageCount < membership.DailyLimit, nil
}

// IncrementUsage 增加用户使用量
func (s *MembershipServiceImpl) IncrementUsage(userID uint) error {
	today := time.Now().Format("2006-01-02")

	// 使用事务保证原子性
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var usage models.UserUsage
	if err := tx.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			usage = models.UserUsage{
				UserID:     userID,
				Date:       time.Now(),
				UsageCount: 1,
			}
			if err := tx.Create(&usage).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	} else {
		usage.UsageCount++
		if err := tx.Save(&usage).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetMembershipInfo 获取会员信息
func (s *MembershipServiceImpl) GetMembershipInfo(userID uint) (*models.Membership, error) {
	var membership models.Membership
	if err := s.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		return nil, err
	}
	return &membership, nil
}
