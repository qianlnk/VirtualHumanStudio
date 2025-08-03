package models

// DigitalHumanTemplate 数字人模版
// 用于管理原始视频、人脸图片、背景图片/视频
// 如果有人脸图片则执行换脸流程，如果有背景则执行换背景流程
// Status: pending/processing/completed/failed

type DigitalHumanTemplate struct {
	BaseModel
	UserID                    uint   `json:"user_id" gorm:"index;not null"`
	Name                      string `json:"name" gorm:"size:100;not null"`
	Description               string `json:"description" gorm:"size:500"`
	VideoURL                  string `json:"video_url" gorm:"size:255;not null"`
	FaceImageURL              string `json:"face_image_url" gorm:"size:255"` // 可选，人脸图片
	BackgroundURL             string `json:"background_url" gorm:"size:255"` // 可选，背景图片或视频
	Status                    string `json:"status" gorm:"default:'pending'"`
	ErrorMsg                  string `json:"error_msg" gorm:"size:500"`                    // 错误信息
	ReplaceFaceVideoURL       string `json:"replace_face_video_url" gorm:"size:255"`       // 可选，替换人脸后的视频
	ReplaceBackgroundVideoURL string `json:"replace_background_video_url" gorm:"size:255"` // 可选，替换背景后的视频
	ResultVideoURL            string `json:"result_video_url" gorm:"size:255"`             // 合成后的视频
}
