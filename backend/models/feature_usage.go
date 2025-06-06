package models

// FeatureType 功能类型
type FeatureType string

const (
	FeatureVoiceClone   FeatureType = "voice_clone"   // 音色克隆
	FeatureTTS          FeatureType = "tts"           // 语音合成
	FeatureASR          FeatureType = "asr"           // 语音识别
	FeatureDigitalHuman FeatureType = "digital_human" // 数字人合成
	FeatureImageProcess FeatureType = "image_process" // 图片处理
)

// FeatureUsage 功能使用量模型
type FeatureUsage struct {
	UserID      uint        `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	Date        string      `json:"date" gorm:"primaryKey;size:10"`
	FeatureType FeatureType `json:"feature_type" gorm:"primaryKey;size:20"`
	UsageCount  int         `json:"usage_count" gorm:"default:0"`
	UsageValue  int         `json:"usage_value" gorm:"default:0"` // 用于记录字数、时长等数值
}

// TableName 设置表名
func (FeatureUsage) TableName() string {
	return "feature_usages"
}
