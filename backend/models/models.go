package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// User 用户模型
type User struct {
	BaseModel
	Username      string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password      string         `json:"-" gorm:"size:100;not null"`
	Email         string         `json:"email" gorm:"size:100"`
	Phone         string         `json:"phone" gorm:"size:20"`
	Role          string         `json:"role" gorm:"size:20;default:'user'"` // admin, user
	Status        int            `json:"status" gorm:"default:1"`            // 0: 禁用, 1: 正常
	LastLoginAt   time.Time      `json:"last_login_at"`
	LastLoginIP   string         `json:"last_login_ip" gorm:"size:50"`
	VoiceClones   []VoiceClone   `json:"voice_clones,omitempty" gorm:"foreignKey:UserID"`
	TTSTasks      []TTSTask      `json:"tts_tasks,omitempty" gorm:"foreignKey:UserID"`
	DigitalHumans []DigitalHuman `json:"digital_humans,omitempty" gorm:"foreignKey:UserID"`
}

// VoiceClone 音色克隆模型
type VoiceClone struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	ModelName   string `json:"model_name" gorm:"size:100;not null"`
	PromptFile  string `json:"prompt_file" gorm:"size:255;not null"` // 文件路径
	PromptText  string `json:"prompt_text" gorm:"size:1000"`
	SpeakerName string `json:"speaker_name" gorm:"size:100;not null"`
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	Result      string `json:"result" gorm:"size:255"`                  // 结果文件路径
	TaskID      string `json:"task_id" gorm:"size:100"`                 // 外部API任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
}

// TTSTask TTS任务模型
type TTSTask struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	Type        string `json:"type" gorm:"size:20;not null"`            // text2speech, speech2text
	InputText   string `json:"input_text" gorm:"type:text"`             // 输入文本
	InputFile   string `json:"input_file" gorm:"size:255"`              // 输入文件路径
	OutputFile  string `json:"output_file" gorm:"size:255"`             // 输出文件路径
	SpeakerName string `json:"speaker_name" gorm:"size:100"`            // 使用的音色名称
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	TaskID      string `json:"task_id" gorm:"size:100"`                 // 外部API任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
}

// DigitalHuman 数字人模型
type DigitalHuman struct {
	BaseModel
	UserID          uint   `json:"user_id" gorm:"index;not null"`
	Name            string `json:"name" gorm:"size:100;not null"`
	Description     string `json:"description" gorm:"size:500"`
	AudioURL        string `json:"audio_url" gorm:"size:255;not null"`      // 音频文件路径
	VideoURL        string `json:"video_url" gorm:"size:255;not null"`      // 视频文件路径
	TaskCode        string `json:"task_code" gorm:"size:100;not null"`      // 任务代码
	Chaofen         int    `json:"chaofen" gorm:"default:0"`                // 超分参数
	WatermarkSwitch int    `json:"watermark_switch" gorm:"default:0"`       // 水印开关
	PN              int    `json:"pn" gorm:"default:1"`                     // PN参数
	Status          string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	ResultURL       string `json:"result_url" gorm:"size:255"`              // 结果文件URL
	ErrorMsg        string `json:"error_msg" gorm:"size:500"`
}

// VoiceLibrary 音色库模型
type VoiceLibrary struct {
	BaseModel
	Name        string `json:"name" gorm:"size:100;not null"`       // 音色名称
	Description string `json:"description" gorm:"size:500"`         // 音色描述
	ModelName   string `json:"model_name" gorm:"size:100;not null"` // 模型名称
	ModelFile   string `json:"model_file" gorm:"size:255;not null"` // 音色模型文件路径
	SampleFile  string `json:"sample_file" gorm:"size:255"`         // 试听音频文件路径
	Type        string `json:"type" gorm:"size:20;not null"`        // original, cloned
	OwnerID     uint   `json:"owner_id" gorm:"index"`               // 所有者ID，如果是克隆音色
	IsPublic    bool   `json:"is_public" gorm:"default:false"`      // 是否公开
}
