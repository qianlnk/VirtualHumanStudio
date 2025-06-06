package config

// ComfyUIConfig ComfyUI配置
type ComfyUIConfig struct {
	// ComfyUI服务器地址
	ServerURL string `json:"comfyui_server_url"`
	// 图片上传接口
	UploadImageAPI string `json:"comfyui_upload_image_api"`
	// 蒙版上传接口
	UploadMaskAPI string `json:"comfyui_upload_mask_api"`
	// 提交任务接口
	PromptAPI string `json:"comfyui_prompt_api"`
	// 查询任务进度接口
	HistoryAPI string `json:"comfyui_history_api"`
	// 取消任务接口
	InterruptAPI string `json:"comfyui_interrupt_api"`
	// 客户端ID
	ClientID string `json:"comfyui_client_id"`
	// 默认模型名称
	DefaultModel string `json:"comfyui_default_model"`
}
