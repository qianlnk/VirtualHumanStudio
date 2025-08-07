package models

import (
	"strings"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/client/promptt"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// ChatSession 聊天会话模型
type ChatSession struct {
	BaseModel
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(50)"`
	Title        string    `json:"title" gorm:"type:varchar(200);not null"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	MessageCount int       `json:"message_count" gorm:"default:0"`
	IsDeleted    bool      `json:"is_deleted" gorm:"default:false"`

	// 关联关系
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID               string    `json:"id" gorm:"primaryKey;type:varchar(50)"`
	SessionID        string    `json:"session_id" gorm:"type:varchar(50);not null;index"`
	UserID           uint      `json:"user_id" gorm:"index;not null"`
	Role             string    `json:"role" gorm:"type:varchar(20);not null"` // user, assistant
	Content          string    `json:"content" gorm:"type:text;not null"`
	ReasoningContent string    `json:"reasoning_content" gorm:"type:text"`
	Model            string    `json:"model" gorm:"type:varchar(100)"`     // AI模型名称
	ImageURL         string    `json:"image_url" gorm:"type:varchar(500)"` // 图像URL
	VideoURL         string    `json:"video_url" gorm:"type:varchar(500)"` // 视频URL
	AudioURL         string    `json:"audio_url" gorm:"type:varchar(500)"` // 音频URL
	FileURL          string    `json:"file_url" gorm:"type:varchar(500)"`  // 通用文件URL
	FileType         string    `json:"file_type" gorm:"type:varchar(50)"`  // 文件类型(image/video/audio/document/other)
	FileName         string    `json:"file_name" gorm:"type:varchar(255)"` // 文件名称
	FileSize         int64     `json:"file_size" gorm:"default:0"`         // 文件大小(bytes)
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Session ChatSession `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

func (msg *ChatMessage) ToChatCompletionMessage(model string) promptt.ChatCompletionMessage {
	chatCompletionMessage := promptt.ChatCompletionMessage{
		ChatCompletionMessage: openai.ChatCompletionMessage{
			Role: msg.Role,
		},
	}

	if msg.Content != "" {
		chatCompletionMessage.MultiContent = append(chatCompletionMessage.MultiContent, promptt.ChatMessagePart{
			ChatMessagePart: openai.ChatMessagePart{
				Type: "text",
				Text: msg.Content,
			},
		})
	}

	if msg.ImageURL != "" {
		chatCompletionMessage.MultiContent = append(chatCompletionMessage.MultiContent, promptt.ChatMessagePart{
			ChatMessagePart: openai.ChatMessagePart{
				Type: "image_url",
				ImageURL: &openai.ChatMessageImageURL{
					URL: msg.ImageURL,
				},
			},
		})
	}

	if msg.VideoURL != "" {
		// gemini的视频服务端只支持了传给图片
		if strings.Contains(model, "gemini") {
			chatCompletionMessage.MultiContent = append(chatCompletionMessage.MultiContent, promptt.ChatMessagePart{
				ChatMessagePart: openai.ChatMessagePart{
					Type: "image_url",
					ImageURL: &openai.ChatMessageImageURL{
						URL: msg.VideoURL,
					},
				},
			})
		} else {
			chatCompletionMessage.MultiContent = append(chatCompletionMessage.MultiContent, promptt.ChatMessagePart{
				ChatMessagePart: openai.ChatMessagePart{
					Type: "video_url",
				},
				VideoURL: &promptt.ChatMessageVideoURL{
					URL: msg.VideoURL,
				},
			})
		}
	}

	return chatCompletionMessage
}

// ModelInfo AI模型信息模型
type ModelInfo struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(50)"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null;unique"`
	Description string    `json:"description" gorm:"type:text"`
	Context     string    `json:"context" gorm:"type:varchar(100)"`
	InputPrice  string    `json:"input_price" gorm:"type:varchar(50)"`
	OutputPrice string    `json:"output_price" gorm:"type:varchar(50)"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ChatRepository 聊天数据访问接口
type ChatRepository interface {
	// 会话相关
	CreateSession(session *ChatSession) error
	GetSessionsByUserID(userID uint) ([]ChatSession, error)
	GetSessionByID(sessionID string) (*ChatSession, error)
	UpdateSession(session *ChatSession) error
	DeleteSession(sessionID string) error

	// 消息相关
	CreateMessage(message *ChatMessage) error
	GetMessagesBySessionID(sessionID string) ([]ChatMessage, error)
	GetMessageByID(messageID string) (*ChatMessage, error)
	UpdateMessage(message *ChatMessage) error
	DeleteMessage(messageID string) error

	// 模型相关
	GetAllModels() ([]ModelInfo, error)
	GetActiveModels() ([]ModelInfo, error)
	CreateModel(model *ModelInfo) error
	UpdateModel(model *ModelInfo) error
	DeleteModel(modelID string) error
}

// ChatRepositoryImpl 聊天数据访问实现
type ChatRepositoryImpl struct {
	db *gorm.DB
}

// NewChatRepository 创建聊天数据访问实例
func NewChatRepository(db *gorm.DB) ChatRepository {
	return &ChatRepositoryImpl{db: db}
}

// CreateSession 创建会话
func (r *ChatRepositoryImpl) CreateSession(session *ChatSession) error {
	return r.db.Create(session).Error
}

// GetSessionsByUserID 获取用户的会话列表
func (r *ChatRepositoryImpl) GetSessionsByUserID(userID uint) ([]ChatSession, error) {
	var sessions []ChatSession
	err := r.db.Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("updated_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetSessionByID 根据ID获取会话
func (r *ChatRepositoryImpl) GetSessionByID(sessionID string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.Where("id = ? AND is_deleted = ?", sessionID, false).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSession 更新会话
func (r *ChatRepositoryImpl) UpdateSession(session *ChatSession) error {
	return r.db.Save(session).Error
}

// DeleteSession 删除会话（软删除）
func (r *ChatRepositoryImpl) DeleteSession(sessionID string) error {
	return r.db.Model(&ChatSession{}).
		Where("id = ?", sessionID).
		Update("is_deleted", true).Error
}

// CreateMessage 创建消息
func (r *ChatRepositoryImpl) CreateMessage(message *ChatMessage) error {
	return r.db.Create(message).Error
}

// GetMessagesBySessionID 获取会话的消息列表
func (r *ChatRepositoryImpl) GetMessagesBySessionID(sessionID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// GetMessageByID 根据ID获取消息
func (r *ChatRepositoryImpl) GetMessageByID(messageID string) (*ChatMessage, error) {
	var message ChatMessage
	err := r.db.Where("id = ?", messageID).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// UpdateMessage 更新消息
func (r *ChatRepositoryImpl) UpdateMessage(message *ChatMessage) error {
	return r.db.Save(message).Error
}

// DeleteMessage 删除消息
func (r *ChatRepositoryImpl) DeleteMessage(messageID string) error {
	return r.db.Delete(&ChatMessage{}, "id = ?", messageID).Error
}

// GetAllModels 获取所有模型
func (r *ChatRepositoryImpl) GetAllModels() ([]ModelInfo, error) {
	var models []ModelInfo
	err := r.db.Order("created_at ASC").Find(&models).Error
	return models, err
}

// GetActiveModels 获取活跃的模型
func (r *ChatRepositoryImpl) GetActiveModels() ([]ModelInfo, error) {
	var models []ModelInfo
	err := r.db.Where("is_active = ?", true).
		Order("created_at ASC").
		Find(&models).Error
	return models, err
}

// CreateModel 创建模型
func (r *ChatRepositoryImpl) CreateModel(model *ModelInfo) error {
	return r.db.Create(model).Error
}

// UpdateModel 更新模型
func (r *ChatRepositoryImpl) UpdateModel(model *ModelInfo) error {
	return r.db.Save(model).Error
}

// DeleteModel 删除模型
func (r *ChatRepositoryImpl) DeleteModel(modelID string) error {
	return r.db.Delete(&ModelInfo{}, "id = ?", modelID).Error
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&ChatSession{},
		&ChatMessage{},
		&ModelInfo{},
	)
}

// InitDefaultModels 初始化默认模型数据
func InitDefaultModels(db *gorm.DB) error {
	defaultModels := []ModelInfo{
		{
			ID:          "dalle-3",
			Name:        "DALL-E 3",
			Description: "OpenAI最新的图像生成模型，擅长创作高质量的艺术作品",
			Context:     "支持复杂提示词",
			InputPrice:  "$0.04 / 1K tokens",
			OutputPrice: "$0.08 / image",
			IsActive:    true,
		},
		{
			ID:          "midjourney",
			Name:        "Midjourney",
			Description: "专注于艺术创作的AI绘画模型，风格独特",
			Context:     "艺术风格丰富",
			InputPrice:  "$0.10 / prompt",
			OutputPrice: "$0.10 / image",
			IsActive:    true,
		},
		{
			ID:          "stable-diffusion",
			Name:        "Stable Diffusion",
			Description: "开源的图像生成模型，支持多种风格和参数调整",
			Context:     "高度可定制",
			InputPrice:  "$0.02 / image",
			OutputPrice: "$0.02 / image",
			IsActive:    true,
		},
		{
			ID:          "claude-3-5-sonnet",
			Name:        "Claude 3.5 Sonnet",
			Description: "Anthropic的智能模型，擅长理解和创作",
			Context:     "200K tokens",
			InputPrice:  "$3.00 / 1M tokens",
			OutputPrice: "$15.00 / 1M tokens",
			IsActive:    true,
		},
	}

	for _, model := range defaultModels {
		var existingModel ModelInfo
		if err := db.Where("id = ?", model.ID).First(&existingModel).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&model).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
