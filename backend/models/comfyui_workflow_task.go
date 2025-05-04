package models

// ComfyUIWorkflowTask 通用ComfyUI工作流任务模型
type ComfyUIWorkflowTask struct {
	ComfyUITask
	// 工作流名称，对应workflow目录下的配置文件名
	WorkflowName string `json:"workflow_name" gorm:"size:100;not null"`
	// 工作流输入参数，JSON格式存储
	InputParams string `json:"input_params" gorm:"type:text"`
	// 工作流输出参数，JSON格式存储
	OutputParams string `json:"output_params" gorm:"type:text"`
}

type InputParam struct {
	Key   string `json:"key"`   // 参数名
	Value string `json:"value"` // 参数值
	Alias string `json:"alias"` // 别名
	Type  string `json:"type"`  // 参数类型
}

const (
	ParamTypeText  = "text"
	ParamTypeImage = "image"
	ParamTypeMask  = "mask"
	ParamTypeVideo = "video"
)
