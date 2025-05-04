package models

// ComfyUITask ComfyUI任务基础模型
type ComfyUITask struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	TaskType    string `json:"task_type" gorm:"size:50;not null"`       // accessory, clothes等
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	TaskID      string `json:"task_id" gorm:"size:100"`                 // ComfyUI任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
}

// Accessory 物品替换任务模型
type Accessory struct {
	ComfyUITask
	ItemImage     string `json:"item_image" gorm:"size:255;not null"`  // 物品白底图路径
	ModelImage    string `json:"model_image" gorm:"size:255;not null"` // 模特图路径
	MaskImage     string `json:"mask_image" gorm:"size:255;not null"`  // 蒙版图路径
	ResultImage   string `json:"result_image" gorm:"size:255"`         // 结果图路径
	Prompt        string `json:"prompt" gorm:"type:text"`              // 提交到ComfyUI的完整prompt
	ItemImageRef  string `json:"item_image_ref" gorm:"size:255"`       // ComfyUI服务器上的物品图引用
	ModelImageRef string `json:"model_image_ref" gorm:"size:255"`      // ComfyUI服务器上的模特图引用
	MaskImageRef  string `json:"mask_image_ref" gorm:"size:255"`       // ComfyUI服务器上的蒙版图引用
}
