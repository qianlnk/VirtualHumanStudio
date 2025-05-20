package services

import (
	"fmt"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MembershipService 会员服务接口
type MembershipService interface {
	// 通用使用限制检查
	// CheckUsageLimit(userID uint) (bool, error)
	// IncrementUsage(userID uint) error
	GetMembershipInfo(userID uint) (*models.Membership, error)

	// 特定功能使用限制检查
	CheckFeatureUsageLimit(userID uint, featureType models.FeatureType) (bool, string, error)
	IncrementFeatureUsage(userID uint, featureType models.FeatureType, usageValue int) error
	GetFeatureUsage(userID uint, featureType models.FeatureType) (*models.FeatureUsage, error)
}

// MembershipServiceImpl 会员服务实现
type MembershipServiceImpl struct {
	DB *gorm.DB
}

// NewMembershipService 创建会员服务
func NewMembershipService(db *gorm.DB) MembershipService {
	// 确保feature_usages表存在
	db.AutoMigrate(&models.FeatureUsage{})
	return &MembershipServiceImpl{DB: db}
}

// // CheckUsageLimit 检查用户是否超出使用限制
// func (s *MembershipServiceImpl) CheckUsageLimit(userID uint) (bool, error) {
// 	var membership models.Membership
// 	if err := s.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
// 		return false, err
// 	}

// 	// 无限制
// 	if membership.DailyLimit < 0 {
// 		return true, nil
// 	}

// 	today := time.Now().Format("2006-01-02")
// 	var usage models.UserUsage
// 	if err := s.DB.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			// 没有今日记录，肯定没超限
// 			return true, nil
// 		}
// 		return false, err
// 	}

// 	// 检查是否超限
// 	return usage.UsageCount < membership.DailyLimit, nil
// }

// // IncrementUsage 增加用户使用量
// func (s *MembershipServiceImpl) IncrementUsage(userID uint) error {
// 	today := time.Now().Format("2006-01-02")

// 	// 定义最大重试次数和初始等待时间
// 	maxRetries := 5
// 	baseWaitTime := 100 * time.Millisecond

// 	// 重试循环
// 	for attempt := 0; attempt < maxRetries; attempt++ {
// 		// 使用事务保证原子性
// 		tx := s.DB.Begin()
// 		defer func() {
// 			if r := recover(); r != nil {
// 				tx.Rollback()
// 			}
// 		}()

// 		var usage models.FeatureUsage
// 		if err := tx.Where("user_id = ? AND date = ?", userID, today).Clauses(clause.Locking{Strength: "UPDATE"}).First(&usage).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				usage = models.FeatureUsage{
// 					UserID:     userID,
// 					Date:       time.Now().Format("2006-01-02"),
// 					UsageCount: 1,
// 				}
// 				if err := tx.Create(&usage).Error; err != nil {
// 					tx.Rollback()
// 					// 检查是否是数据库锁定错误
// 					if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
// 						// 计算等待时间（指数退避）
// 						waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
// 						fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
// 						time.Sleep(waitTime)
// 						continue
// 					}
// 					return err
// 				}
// 			} else {
// 				tx.Rollback()
// 				// 检查是否是数据库锁定错误
// 				if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
// 					// 计算等待时间（指数退避）
// 					waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
// 					fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
// 					time.Sleep(waitTime)
// 					continue
// 				}
// 				return err
// 			}
// 		} else {
// 			usage.UsageCount++
// 			if err := tx.Save(&usage).Error; err != nil {
// 				tx.Rollback()
// 				// 检查是否是数据库锁定错误
// 				if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
// 					// 计算等待时间（指数退避）
// 					waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
// 					fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
// 					time.Sleep(waitTime)
// 					continue
// 				}
// 				return err
// 			}
// 		}

// 		// 提交事务
// 		if err := tx.Commit().Error; err != nil {
// 			// 检查是否是数据库锁定错误
// 			if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
// 				// 计算等待时间（指数退避）
// 				waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
// 				fmt.Printf("提交事务时数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
// 				time.Sleep(waitTime)
// 				continue
// 			}
// 			return err
// 		}

// 		// 如果执行到这里，说明事务成功提交，可以退出循环
// 		return nil
// 	}

// 	// 如果所有重试都失败了
// 	return fmt.Errorf("增加用户使用量失败：数据库锁定，已重试 %d 次", maxRetries)
// }

// GetMembershipInfo 获取会员信息
func (s *MembershipServiceImpl) GetMembershipInfo(userID uint) (*models.Membership, error) {
	var membership models.Membership
	if err := s.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		return nil, err
	}
	return &membership, nil
}

// CheckFeatureUsageLimit 检查特定功能的使用限制
func (s *MembershipServiceImpl) CheckFeatureUsageLimit(userID uint, featureType models.FeatureType) (bool, string, error) {
	// 获取会员信息
	fmt.Println("============", featureType)
	var membership models.Membership
	if err := s.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		return false, "获取会员信息失败", err
	}

	fmt.Println("会员信息", utils.ToJSONString(membership))

	// 获取会员计划详情
	var plan models.MembershipPlan
	if err := s.DB.Where("level = ?", membership.Level).First(&plan).Error; err != nil {
		return false, "获取会员计划失败", err
	}

	fmt.Println("会员计划", utils.ToJSONString(plan))

	// 获取今日使用量
	today := time.Now().Format("2006-01-02")
	var usage models.FeatureUsage
	result := s.DB.Where("user_id = ? AND date = ? AND feature_type = ?", userID, today, featureType).First(&usage)
	fmt.Println("今日用量", utils.ToJSONString(usage))
	// 根据功能类型检查限制
	switch featureType {
	case models.FeatureVoiceClone:
		// 无限制
		if plan.VoiceClonePerDay < 0 {
			return true, "", nil
		}
		// 没有记录或未超限
		if result.Error == gorm.ErrRecordNotFound || usage.UsageCount < plan.VoiceClonePerDay {
			return true, "", nil
		}
		return false, "您今日的音色克隆次数已达上限，请升级会员或明日再试", nil

	case models.FeatureTTS:
		// 无限制
		if plan.TTSWordsPerDay < 0 {
			return true, "", nil
		}
		// 没有记录或未超限
		if result.Error == gorm.ErrRecordNotFound || usage.UsageValue < plan.TTSWordsPerDay {
			return true, "", nil
		}
		return false, "您今日的语音合成字数已达上限，请升级会员或明日再试", nil

	case models.FeatureASR:
		// 无限制
		if plan.ASRTimesPerDay < 0 {
			return true, "", nil
		}
		// 没有记录或未超限
		if result.Error == gorm.ErrRecordNotFound || usage.UsageCount < plan.ASRTimesPerDay {
			return true, "", nil
		}
		return false, "您今日的语音识别次数已达上限，请升级会员或明日再试", nil

	case models.FeatureDigitalHuman:
		// 无限制
		if plan.DigitalHumanPerDay < 0 {
			return true, "", nil
		}
		// 没有记录或未超限
		if result.Error == gorm.ErrRecordNotFound || usage.UsageCount < plan.DigitalHumanPerDay {
			return true, "", nil
		}
		return false, "您今日的数字人合成次数已达上限，请升级会员或明日再试", nil

	case models.FeatureImageProcess:
		// 无限制
		if plan.ImageProcessPerDay < 0 {
			return true, "", nil
		}
		// 没有记录或未超限
		if result.Error == gorm.ErrRecordNotFound || usage.UsageCount < plan.ImageProcessPerDay {
			return true, "", nil
		}
		return false, "您今日的图片处理次数已达上限，请升级会员或明日再试", nil

	default:
		return true, "", nil
	}
}

// IncrementFeatureUsage 增加特定功能的使用量
func (s *MembershipServiceImpl) IncrementFeatureUsage(userID uint, featureType models.FeatureType, usageValue int) error {
	today := time.Now().Format("2006-01-02")

	// 定义最大重试次数和初始等待时间
	maxRetries := 5
	baseWaitTime := 100 * time.Millisecond

	// 重试循环
	for attempt := 0; attempt < maxRetries; attempt++ {
		// 使用事务保证原子性
		tx := s.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		var usage models.FeatureUsage
		if err := tx.Where("user_id = ? AND date = ? AND feature_type = ?", userID, today, featureType).Clauses(clause.Locking{Strength: "UPDATE"}).First(&usage).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新记录
				usage = models.FeatureUsage{
					UserID:      userID,
					Date:        time.Now().Format("2006-01-02"),
					FeatureType: featureType,
					UsageCount:  1,
					UsageValue:  usageValue,
				}
				if err := tx.Create(&usage).Error; err != nil {
					tx.Rollback()
					// 检查是否是数据库锁定错误
					if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
						// 计算等待时间（指数退避）
						waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
						fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
						time.Sleep(waitTime)
						continue
					}
					return err
				}
			} else {
				tx.Rollback()
				// 检查是否是数据库锁定错误
				if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
					// 计算等待时间（指数退避）
					waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
					fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
					time.Sleep(waitTime)
					continue
				}
				return err
			}
		} else {
			// 更新现有记录
			usage.UsageCount++
			usage.UsageValue += usageValue
			if err := tx.Save(&usage).Error; err != nil {
				fmt.Println("更新现有记录", err)
				tx.Rollback()
				// 检查是否是数据库锁定错误
				if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
					// 计算等待时间（指数退避）
					waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
					fmt.Printf("数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
					time.Sleep(waitTime)
					continue
				}
				return err
			}
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			// 检查是否是数据库锁定错误
			if err.Error() == "database is locked" || err.Error() == "database is locked (5) (SQLITE_BUSY)" {
				// 计算等待时间（指数退避）
				waitTime := baseWaitTime * time.Duration(1<<uint(attempt))
				fmt.Printf("提交事务时数据库锁定，等待 %v 后重试 (尝试 %d/%d)\n", waitTime, attempt+1, maxRetries)
				time.Sleep(waitTime)
				continue
			}
			return err
		}

		// 如果执行到这里，说明事务成功提交，可以退出循环
		return nil
	}

	// 如果所有重试都失败了
	return fmt.Errorf("增加功能使用量失败：数据库锁定，已重试 %d 次", maxRetries)
}

// GetFeatureUsage 获取特定功能的使用量
func (s *MembershipServiceImpl) GetFeatureUsage(userID uint, featureType models.FeatureType) (*models.FeatureUsage, error) {
	today := time.Now().Format("2006-01-02")
	var usage models.FeatureUsage

	if err := s.DB.Where("user_id = ? AND date = ? AND feature_type = ?", userID, today, featureType).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回空记录
			return &models.FeatureUsage{
				UserID:      userID,
				Date:        time.Now().Format("2006-01-02"),
				FeatureType: featureType,
				UsageCount:  0,
				UsageValue:  0,
			}, nil
		}
		return nil, err
	}

	return &usage, nil
}
