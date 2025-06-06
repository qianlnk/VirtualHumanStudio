package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/redis"
)

const (
	// 队列键名
	VoiceCloneMemberQueueKey    = "voice_clone:queue:member"     // 会员队列
	VoiceCloneNonMemberQueueKey = "voice_clone:queue:non_member" // 非会员队列
	VoiceCloneProcessingSetKey  = "voice_clone:processing"       // 处理中的任务集合
	VoiceCloneQueueInfoKey      = "voice_clone:queue_info"       // 队列信息哈希表
)

// QueueItem 队列项
type QueueItem struct {
	ID        uint      `json:"id"`         // 任务ID
	UserID    uint      `json:"user_id"`    // 用户ID
	IsMember  bool      `json:"is_member"`  // 是否为会员
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// QueueInfo 队列信息
type QueueInfo struct {
	MemberCount     int `json:"member_count"`     // 会员队列长度
	NonMemberCount  int `json:"non_member_count"` // 非会员队列长度
	ProcessingCount int `json:"processing_count"` // 处理中的任务数
	MaxProcessing   int `json:"max_processing"`   // 最大同时处理数
}

func memberQueueKey(featureType string) string {
	return fmt.Sprintf("membership:%s:queue", featureType)
}
func nonMemberQueueKey(featureType string) string {
	return fmt.Sprintf("non_membership:%s:queue", featureType)
}
func processingSetKey(featureType string) string {
	return fmt.Sprintf("processing:%s", featureType)
}
func queueInfoKey(featureType string) string {
	return fmt.Sprintf("queue_info:%s", featureType)
}

// 音色克隆队列服务
// AddToVoiceCloneQueue 添加任务到音色克隆队列
func AddToVoiceCloneQueue(ctx context.Context, taskID uint, userID uint, isMember bool) error {
	return AddToQueue(ctx, string(models.FeatureVoiceClone), taskID, userID, isMember)
}

// GetVoiceCloneQueueInfo 获取音色克隆队列信息
func GetVoiceCloneQueueInfo(ctx context.Context) (*QueueInfo, error) {
	return GetQueueInfo(ctx, string(models.FeatureVoiceClone))
}

// StartVoiceCloneQueueConsumer 启动音色克隆队列消费者
func StartVoiceCloneQueueConsumer(ctx context.Context, processTask func(context.Context, uint) error) {
	StartQueueConsumer(ctx, string(models.FeatureVoiceClone), processTask)
}

// GetVoiceCloneQueuePosition 获取音色克隆队列中任务的位置
func GetVoiceCloneQueuePosition(ctx context.Context, taskID uint) (int, string, error) {
	return GetQueuePosition(ctx, string(models.FeatureVoiceClone), taskID)
}

// 语音合成队列服务
// AddToTTSQueue 添加任务到语音合成队列
func AddToTTSQueue(ctx context.Context, taskID uint, userID uint, isMember bool) error {
	return AddToQueue(ctx, string(models.FeatureTTS), taskID, userID, isMember)
}

// GetTTSQueueInfo 获取语音合成队列信息
func GetTTSQueueInfo(ctx context.Context) (*QueueInfo, error) {
	return GetQueueInfo(ctx, string(models.FeatureTTS))
}

// StartTTSQueueConsumer 启动语音合成队列消费者
func StartTTSQueueConsumer(ctx context.Context, processTask func(context.Context, uint) error) {
	StartQueueConsumer(ctx, string(models.FeatureTTS), processTask)
}

// GetTTSQueuePosition 获取语音合成队列中任务的位置
func GetTTSQueuePosition(ctx context.Context, taskID uint) (int, string, error) {
	return GetQueuePosition(ctx, string(models.FeatureTTS), taskID)
}

// 语音识别队列服务
// AddToASRQueue 添加任务到语音识别队列
func AddToASRQueue(ctx context.Context, taskID uint, userID uint, isMember bool) error {
	return AddToQueue(ctx, string(models.FeatureASR), taskID, userID, isMember)
}

// GetASRQueueInfo 获取语音识别队列信息
func GetASRQueueInfo(ctx context.Context) (*QueueInfo, error) {
	return GetQueueInfo(ctx, string(models.FeatureASR))
}

// StartASRQueueConsumer 启动语音识别队列消费者
func StartASRQueueConsumer(ctx context.Context, processTask func(context.Context, uint) error) {
	StartQueueConsumer(ctx, string(models.FeatureASR), processTask)
}

// GetASRQueuePosition 获取语音识别队列中任务的位置
func GetASRQueuePosition(ctx context.Context, taskID uint) (int, string, error) {
	return GetQueuePosition(ctx, string(models.FeatureASR), taskID)
}

// 数字人合成队列服务
// AddToDigitalHumanQueue 添加任务到数字人合成队列
func AddToDigitalHumanQueue(ctx context.Context, taskID uint, userID uint, isMember bool) error {
	return AddToQueue(ctx, string(models.FeatureDigitalHuman), taskID, userID, isMember)
}

// GetDigitalHumanQueueInfo 获取数字人合成队列信息
func GetDigitalHumanQueueInfo(ctx context.Context) (*QueueInfo, error) {
	return GetQueueInfo(ctx, string(models.FeatureDigitalHuman))
}

// StartDigitalHumanQueueConsumer 启动数字人合成队列消费者
func StartDigitalHumanQueueConsumer(ctx context.Context, processTask func(context.Context, uint) error) {
	StartQueueConsumer(ctx, string(models.FeatureDigitalHuman), processTask)
}

// GetDigitalHumanQueuePosition 获取数字人合成队列中任务的位置
func GetDigitalHumanQueuePosition(ctx context.Context, taskID uint) (int, string, error) {
	return GetQueuePosition(ctx, string(models.FeatureDigitalHuman), taskID)
}

// 图片处理队列服务
// AddToImageProcessQueue 添加任务到图片处理队列
func AddToImageProcessQueue(ctx context.Context, taskID uint, userID uint, isMember bool) error {
	return AddToQueue(ctx, string(models.FeatureImageProcess), taskID, userID, isMember)
}

// GetImageProcessQueueInfo 获取图片处理队列信息
func GetImageProcessQueueInfo(ctx context.Context) (*QueueInfo, error) {
	return GetQueueInfo(ctx, string(models.FeatureImageProcess))
}

// StartImageProcessQueueConsumer 启动图片处理队列消费者
func StartImageProcessQueueConsumer(ctx context.Context, processTask func(context.Context, uint) error) {
	StartQueueConsumer(ctx, string(models.FeatureImageProcess), processTask)
}

// GetImageProcessQueuePosition 获取图片处理队列中任务的位置
func GetImageProcessQueuePosition(ctx context.Context, taskID uint) (int, string, error) {
	return GetQueuePosition(ctx, string(models.FeatureImageProcess), taskID)
}

// AddToVoiceCloneQueue 添加任务到音色克隆队列
func AddToQueue(ctx context.Context, featureType string, taskID uint, userID uint, isMember bool) error {
	redisClient := redis.GetRedisClient()

	// 创建队列项
	item := QueueItem{
		ID:        taskID,
		UserID:    userID,
		IsMember:  isMember,
		CreatedAt: time.Now(),
	}

	// 序列化队列项
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("序列化队列项失败: %v", err)
	}

	// 根据会员状态选择队列
	queueKey := nonMemberQueueKey(featureType)
	if isMember {
		queueKey = memberQueueKey(featureType)
	}

	// 添加到队列
	err = redisClient.RPush(ctx, queueKey, string(itemJSON)).Err()
	if err != nil {
		return fmt.Errorf("添加到队列失败: %v", err)
	}

	// 更新队列信息
	updateQueueInfo(ctx, featureType, redisClient)

	return nil
}

func GetQueueInfo(ctx context.Context, featureType string) (*QueueInfo, error) {
	redisClient := redis.GetRedisClient()

	// 获取队列长度
	memberCount, err := redisClient.LLen(ctx, memberQueueKey(featureType)).Result()
	if err != nil {
		return nil, fmt.Errorf("获取会员队列长度失败: %v", err)
	}

	nonMemberCount, err := redisClient.LLen(ctx, nonMemberQueueKey(featureType)).Result()
	if err != nil {
		return nil, fmt.Errorf("获取非会员队列长度失败: %v", err)
	}

	// 获取处理中的任务数
	processingCount, err := redisClient.SCard(ctx, processingSetKey(featureType)).Result()
	if err != nil {
		return nil, fmt.Errorf("获取处理中任务数失败: %v", err)
	}

	maxProcessing := 1
	switch featureType {
	case string(models.FeatureVoiceClone):
		maxProcessing = 1
	case string(models.FeatureTTS):
		maxProcessing = 3
	case string(models.FeatureASR):
		maxProcessing = 5
	case string(models.FeatureDigitalHuman):
		maxProcessing = 1
	case string(models.FeatureImageProcess):
		maxProcessing = 1
	default:
		maxProcessing = 1
	}

	// 返回队列信息
	return &QueueInfo{
		MemberCount:     int(memberCount),
		NonMemberCount:  int(nonMemberCount),
		ProcessingCount: int(processingCount),
		MaxProcessing:   maxProcessing,
	}, nil
}

// GetQueuePosition 获取任务在队列中的位置
func GetQueuePosition(ctx context.Context, featureType string, taskID uint) (int, string, error) {
	redisClient := redis.GetRedisClient()

	// 检查是否在处理中
	isProcessing, err := redisClient.SIsMember(ctx, processingSetKey(featureType), fmt.Sprintf("%d", taskID)).Result()
	if err != nil {
		return -1, "", fmt.Errorf("检查任务状态失败: %v", err)
	}

	if isProcessing {
		return 0, "processing", nil
	}

	// 检查会员队列
	memberItems, err := redisClient.LRange(ctx, memberQueueKey(featureType), 0, -1).Result()
	if err != nil {
		return -1, "", fmt.Errorf("获取会员队列失败: %v", err)
	}

	for i, itemStr := range memberItems {
		var item QueueItem
		if err := json.Unmarshal([]byte(itemStr), &item); err != nil {
			continue
		}

		if item.ID == taskID {
			return i + 1, "member", nil
		}
	}

	// 检查非会员队列
	nonMemberItems, err := redisClient.LRange(ctx, nonMemberQueueKey(featureType), 0, -1).Result()
	if err != nil {
		return -1, "", fmt.Errorf("获取非会员队列失败: %v", err)
	}

	for i, itemStr := range nonMemberItems {
		var item QueueItem
		if err := json.Unmarshal([]byte(itemStr), &item); err != nil {
			continue
		}

		if item.ID == taskID {
			// 非会员队列位置需要加上会员队列长度
			return i + 1 + len(memberItems), "non_member", nil
		}
	}

	return -1, "", nil
}

func StartQueueConsumer(ctx context.Context, featureType string, processTask func(context.Context, uint) error) {
	redisClient := redis.GetRedisClient()

	go func() {
		for {
			// 检查是否可以处理更多任务
			info, err := GetQueueInfo(ctx, featureType)
			if err != nil {
				log.Printf("获取队列信息失败: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}
			// log.Printf("%s队列信息: %+v\n", featureType, info)
			if info.ProcessingCount >= info.MaxProcessing {
				// 已达到最大处理数，等待一段时间后重试
				time.Sleep(5 * time.Second)
				continue
			}

			// 优先从会员队列获取任务
			var queueKey string
			if info.MemberCount > 0 {
				queueKey = memberQueueKey(featureType)
			} else if info.NonMemberCount > 0 {
				queueKey = nonMemberQueueKey(featureType)
			} else {
				// 队列为空，等待一段时间后重试
				time.Sleep(5 * time.Second)
				continue
			}

			// 从队列中获取任务
			itemStr, err := redisClient.LPop(ctx, queueKey).Result()
			if err != nil {
				if err != goredis.Nil {
					log.Printf("从队列获取任务失败: %v\n", err)
				}
				time.Sleep(5 * time.Second)
				continue
			}

			// 解析队列项
			var item QueueItem
			if err := json.Unmarshal([]byte(itemStr), &item); err != nil {
				log.Printf("解析队列项失败: %v\n", err)
				continue
			}

			// 添加到处理中集合
			err = redisClient.SAdd(ctx, processingSetKey(featureType), fmt.Sprintf("%d", item.ID)).Err()
			if err != nil {
				log.Printf("添加到处理中集合失败: %v\n", err)
				// 将任务放回队列
				redisClient.RPush(ctx, queueKey, itemStr)
				time.Sleep(5 * time.Second)
				continue
			}

			// 更新队列信息
			updateQueueInfo(ctx, featureType, redisClient)

			// 异步处理任务
			go func(taskID uint) {
				defer func() {
					// 处理完成后从处理中集合移除
					redisClient.SRem(ctx, processingSetKey(featureType), fmt.Sprintf("%d", taskID))
					// 更新队列信息
					updateQueueInfo(ctx, featureType, redisClient)
				}()

				// 处理任务
				if err := processTask(ctx, taskID); err != nil {
					log.Printf("处理任务 %d 失败: %v\n", taskID, err)
				}
			}(item.ID)
		}
	}()
}

// 更新队列信息
func updateQueueInfo(ctx context.Context, featureType string, redisClient *goredis.Client) {
	// 获取队列长度
	memberCount, err := redisClient.LLen(ctx, memberQueueKey(featureType)).Result()
	if err != nil {
		log.Printf("获取会员队列长度失败: %v\n", err)
		return
	}

	nonMemberCount, err := redisClient.LLen(ctx, nonMemberQueueKey(featureType)).Result()
	if err != nil {
		log.Printf("获取非会员队列长度失败: %v\n", err)
		return
	}

	// 获取处理中的任务数
	processingCount, err := redisClient.SCard(ctx, processingSetKey(featureType)).Result()
	if err != nil {
		log.Printf("获取处理中任务数失败: %v\n", err)
		return
	}

	// 更新队列信息
	info := QueueInfo{
		MemberCount:     int(memberCount),
		NonMemberCount:  int(nonMemberCount),
		ProcessingCount: int(processingCount),
		MaxProcessing:   5, // 最大同时处理5个任务
	}

	// 序列化队列信息
	infoJSON, err := json.Marshal(info)
	if err != nil {
		log.Printf("序列化队列信息失败: %v\n", err)
		return
	}

	// 保存队列信息
	err = redisClient.Set(ctx, queueInfoKey(featureType), string(infoJSON), 0).Err()
	if err != nil {
		log.Printf("保存队列信息失败: %v\n", err)
		return
	}
}
