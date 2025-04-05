package models

// ASRTask 语音识别任务模型
type ASRTask struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	Model       string `json:"model" gorm:"size:100;not null"`          // ASR模型名称
	InputFile   string `json:"input_file" gorm:"size:255"`              // 输入音频文件路径
	OutputText  string `json:"output_text" gorm:"type:text"`            // 识别结果文本
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	TaskID      string `json:"task_id" gorm:"size:100"`                 // 外部API任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
}
