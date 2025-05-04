package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	ServerPort int    `json:"server_port"`
	ServerHost string `json:"server_host"`
	JWTSecret  string `json:"jwt_secret"`

	// 数据库配置
	DBType     string `json:"db_type"` // sqlite 或 mysql
	SQLitePath string `json:"sqlite_path"`

	// MySQL配置
	MySQLHost string `json:"mysql_host"`
	MySQLPort int    `json:"mysql_port"`
	MySQLUser string `json:"mysql_user"`
	MySQLPass string `json:"mysql_pass"`
	MySQLDB   string `json:"mysql_db"`

	// Redis配置
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// 文件存储配置
	DataDir   string `json:"data_dir"`   // 基础数据目录
	UploadDir string `json:"upload_dir"` // 上传文件存储目录
	VoiceDir  string `json:"voice_dir"`  // 音色文件存储目录
	AudioDir  string `json:"audio_dir"`  // 音频文件存储目录
	VideoDir  string `json:"video_dir"`  // 视频文件存储目录

	// API配置
	FileUploadAPI      string `json:"file_upload_api"`      // 文件上传服务API
	FileDownloadAPI    string `json:"file_download_api"`    // 文件下载服务API
	FileServerBaseURL  string `json:"file_server_base_url"` // 文件服务器基础URL
	FileViewURL        string `json:"file_view_url"`        // 文件查看URL
	VoiceCloneAPI      string `json:"voice_clone_api"`
	TTSAPI             string `json:"tts_api"`
	DigitalHumanAPI    string `json:"digital_human_api"`
	DigitalHumanQuery  string `json:"digital_human_query"`
	DigitalHumanResult string `json:"digital_human_result"`
	ASRAPI             string `json:"asr_api"` // 语音识别服务URL

	// ComfyUI配置
	ComfyUIConf ComfyUIConfig `json:"comfyui"`

	// 域名
	Domain string `json:"domain"` // 域名

	OfficialAPI    string               `json:"official_api"`    // 官方API
	OfficialKey    string               `json:"official_key"`    // 官方Key
	OfficialVoices map[string]TTSTimbre `json:"official_voices"` // 官方音色列表

	Promptt *Promptt `json:"promptt"` // Promptt配置
}

type Promptt struct {
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	ChatPath string `json:"chat_path"`
	TTSPath  string `json:"tts_path"`
	ASRPath  string `json:"asr_path"`
}

type TTSTimbre struct {
	ID          string `json:"id"`          // ID, 用于系统内唯一标识
	Model       string `json:"model"`       // 模型，不对外
	TimbreID    string `json:"timbre_id"`   // 音色ID，不对外
	Name        string `json:"name"`        // 名称，不对外
	Description string `json:"description"` // 描述
	Alias       string `json:"alias"`       // 别名
	Gender      string `json:"gender"`      // 性别
	CreatedAt   string `json:"created_at"`  // 创建时间
	SampleFile  string `json:"sample_file"` // 示例文件
}

// AppConfig 全局配置实例
var AppConfig Config

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) error {
	// 设置默认配置
	AppConfig = Config{
		ServerPort: 8080,
		ServerHost: "0.0.0.0",
		JWTSecret:  "virtual_human_studio_secret_key",

		DBType:     "sqlite",
		SQLitePath: "./data/vhs.db",

		MySQLHost: "localhost",
		MySQLPort: 3306,
		MySQLUser: "root",
		MySQLPass: "",
		MySQLDB:   "virtual_human_studio",

		RedisHost:     "localhost",
		RedisPort:     6379,
		RedisPassword: "",
		RedisDB:       0,

		DataDir:   "./data",
		UploadDir: "uploads",
		VoiceDir:  "voices",
		VideoDir:  "videos",

		FileUploadAPI:      "https://aigc-ops-test.skyengine.com.cn/v1/file/upload2path",
		FileDownloadAPI:    "https://aigc-ops-test.skyengine.com.cn/v1/file/view",
		FileServerBaseURL:  "https://aigc-ops-test.skyengine.com.cn",
		VoiceCloneAPI:      "https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/cosyvoice2-05b:8080/api/voice/clone",
		TTSAPI:             "https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/cosyvoice2-05b:8080/api/tts", // 假设的TTS API
		DigitalHumanAPI:    "https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/heygem-f2f:8383/easy/submit",
		DigitalHumanQuery:  "https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/heygem-f2f:8383/easy/query",
		DigitalHumanResult: "http://10.64.24.249:8383/easy/result", // 假设的结果获取API

		ComfyUIConf: ComfyUIConfig{
			ServerURL:      "https://aigc-ops-test.skyengine.com.cn/v1/model/comfyui/workflow-comfyui",
			UploadImageAPI: "/api/upload/image",
			UploadMaskAPI:  "/api/upload/mask",
			PromptAPI:      "/api/prompt",
			HistoryAPI:     "/api/history",
			InterruptAPI:   "/api/interrupt",
			ClientID:       "virtual_human_studio",
			DefaultModel:   "ghostxl_v10BakedVAE.safetensors",
		},

		OfficialVoices: map[string]TTSTimbre{
			"voice_001": {ID: "voice_001", Model: "ktian-speech_1", TimbreID: "ktian-speech_wuzhao_test_3", Name: "趣小侠", Alias: "灵动侠客", Gender: "女"},
			"voice_002": {ID: "voice_002", Model: "ktian-speech_1", TimbreID: "ktian-speech_Female_ shangshen_platform", Name: "趣小商", Alias: "商务精英", Gender: "女"},
			"voice_003": {ID: "voice_003", Model: "ktian-speech_1", TimbreID: "ktian-speech_guimi0220_hificloneck14", Name: "趣小蜜", Alias: "甜心密友", Gender: "女"},
			"voice_004": {ID: "voice_004", Model: "ktian-speech_1", TimbreID: "ktian-speech_Female_shaonv_platform", Name: "趣小仙", Alias: "仙气少女", Gender: "女"},
			"voice_005": {ID: "voice_005", Model: "ktian-speech_1", TimbreID: "ktian-speech_female-shaonv-jingpin", Name: "趣小柔", Alias: "温柔知心", Gender: "女"},
			"voice_006": {ID: "voice_006", Model: "ktian-speech_1", TimbreID: "ktian-speech_female-chengshu-jingpin", Name: "趣小影", Alias: "优雅成熟", Gender: "女"},
			"voice_007": {ID: "voice_007", Model: "ktian-speech_1", TimbreID: "ktian-speech_Podcast_girl_platform", Name: "趣小琴", Alias: "音乐达人", Gender: "女"},
			"voice_008": {ID: "voice_008", Model: "ktian-speech_1", TimbreID: "ktian-speech_wuzhao_test_5", Name: "趣小无", Alias: "空灵仙子", Gender: "女"},
			"voice_009": {ID: "voice_009", Model: "ktian-speech_1", TimbreID: "ktian-speech_Female_yaoyao_platform", Name: "趣小瑶", Alias: "活力青春", Gender: "女"},
			"voice_010": {ID: "voice_010", Model: "ktian-speech_1", TimbreID: "ktian-speech_guimi0220_hificloneck20", Name: "趣小清", Alias: "清新脱俗", Gender: "女"},
			"voice_011": {ID: "voice_011", Model: "ktian-speech_1", TimbreID: "ktian-speech_wuzhao_test_4", Name: "趣小风", Alias: "飘逸风姿", Gender: "女"},
			"voice_012": {ID: "voice_012", Model: "ktian-speech_1", TimbreID: "ktian-speech_female-tianmei-jingpin", Name: "趣小甜", Alias: "甜美可人", Gender: "女"},
			"voice_013": {ID: "voice_013", Model: "ktian-speech_1", TimbreID: "ktian-speech_guimi0220_hificloneck15", Name: "趣小纯", Alias: "纯真天使", Gender: "女"},
			"voice_014": {ID: "voice_014", Model: "ktian-speech_1", TimbreID: "ktian-speech_wuzhao_test_1", Name: "趣小初", Alias: "初心少女", Gender: "女"},
			"voice_015": {ID: "voice_015", Model: "ktian-speech_1", TimbreID: "ktian-speech_Female_murong_platform", Name: "趣小蓉", Alias: "温婉淑女", Gender: "女"},
			"voice_016": {ID: "voice_016", Model: "ktian-speech_1", TimbreID: "ktian-speech_Female_xiaomo_platform", Name: "趣小墨", Alias: "墨韵书香", Gender: "女"},
			"voice_017": {ID: "voice_017", Model: "ktian-speech_1", TimbreID: "ktian-speech_wuzhao_test_2", Name: "趣小义", Alias: "侠义豪情", Gender: "女"},
			"voice_018": {ID: "voice_018", Model: "ktian-speech_1", TimbreID: "ktian-speech_female-yujie-jingpin", Name: "趣小玉", Alias: "玉质兰心", Gender: "女"},
			"voice_019": {ID: "voice_019", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_61_speech02", Name: "趣小灵", Alias: "灵秀佳人", Gender: "女"},
			"voice_020": {ID: "voice_020", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_71_speech02", Name: "趣小寒", Alias: "寒梅傲雪", Gender: "女"},
			"voice_021": {ID: "voice_021", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_90_speech02", Name: "趣小南", Alias: "南国风情", Gender: "女"},
			"voice_022": {ID: "voice_022", Model: "ktian-speech_1", TimbreID: "ktian-speech_voice69", Name: "趣小飞", Alias: "飞扬青春", Gender: "女"},
			"voice_023": {ID: "voice_023", Model: "ktian-speech_1", TimbreID: "ktian-speech_male-qn-badao-jingpin", Name: "趣小霸", Alias: "霸气总裁", Gender: "男"},
			"voice_024": {ID: "voice_024", Model: "ktian-speech_1", TimbreID: "ktian-speech_Boyan_new_platform", Name: "趣小博", Alias: "博学多才", Gender: "男"},
			"voice_025": {ID: "voice_025", Model: "ktian-speech_1", TimbreID: "ktian-speech_Male_botong_platform", Name: "趣小云", Alias: "云淡风轻", Gender: "男"},
			"voice_026": {ID: "voice_026", Model: "ktian-speech_1", TimbreID: "ktian-speech_male-qn-jingying-jingpin", Name: "趣小鹰", Alias: "鹰击长空", Gender: "男"},
			"voice_027": {ID: "voice_027", Model: "ktian-speech_1", TimbreID: "ktian-speech_male-qn-daxuesheng-jingpin", Name: "趣小儒", Alias: "儒雅书生", Gender: "男"},
			"voice_028": {ID: "voice_028", Model: "ktian-speech_1", TimbreID: "ktian-speech_Male_zhaoyi_platform", Name: "趣小昭", Alias: "昭然明志", Gender: "男"},
			"voice_029": {ID: "voice_029", Model: "ktian-speech_1", TimbreID: "ktian-speech_Xiaoyi_mix_platform", Name: "趣小逸", Alias: "逸群之才", Gender: "男"},
			"voice_030": {ID: "voice_030", Model: "ktian-speech_1", TimbreID: "ktian-speech_Male_zeyang_platform", Name: "趣小阳", Alias: "阳光少年", Gender: "男"},
			"voice_031": {ID: "voice_031", Model: "ktian-speech_1", TimbreID: "ktian-speech_male-qn-qingse-jingpin", Name: "趣小青", Alias: "青云直上", Gender: "男"},
			"voice_032": {ID: "voice_032", Model: "ktian-speech_1", TimbreID: "ktian-speech_Male_kongchen_platform", Name: "趣小空", Alias: "空谷幽兰", Gender: "男"},
			"voice_033": {ID: "voice_033", Model: "ktian-speech_1", TimbreID: "ktian-speech_Bingjiao_zongcai_platform", Name: "趣小冰", Alias: "冰雪聪明", Gender: "男"},
			"voice_034": {ID: "voice_034", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_62_speech02", Name: "趣小安", Alias: "安然如山", Gender: "男"},
			"voice_035": {ID: "voice_035", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_63_speech02", Name: "趣小新", Alias: "新锐先锋", Gender: "男"},
			"voice_036": {ID: "voice_036", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_64_speech02", Name: "趣小然", Alias: "从容淡定", Gender: "男"},
			"voice_037": {ID: "voice_037", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_65_speech02", Name: "趣小远", Alias: "远见卓识", Gender: "男"},
			"voice_038": {ID: "voice_038", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_67_speech02", Name: "趣小岳", Alias: "岳立巍峨", Gender: "男"},
			"voice_039": {ID: "voice_039", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_73_speech02", Name: "趣小光", Alias: "光明磊落", Gender: "男"},
			"voice_040": {ID: "voice_040", Model: "ktian-speech_1", TimbreID: "ktian-speech_minimax_soulmate_voice_id_77_speech02", Name: "趣小锋", Alias: "锋芒毕露", Gender: "男"},
			"voice_041": {ID: "voice_041", Model: "ktian-speech_1", TimbreID: "ktian-speech_62", Name: "趣小尘", Alias: "尘外仙人", Gender: "男"},
		},
	}

	// 如果配置文件存在，则从文件加载配置
	if configPath != "" {
		// 确保配置文件目录存在
		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0755)
			if err != nil {
				return fmt.Errorf("无法创建配置目录: %v", err)
			}
		}

		// 检查配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置文件
			configJSON, err := json.MarshalIndent(AppConfig, "", "  ")
			if err != nil {
				return fmt.Errorf("无法序列化默认配置: %v", err)
			}

			err = os.WriteFile(configPath, configJSON, 0644)
			if err != nil {
				return fmt.Errorf("无法写入默认配置文件: %v", err)
			}

			fmt.Printf("已创建默认配置文件: %s\n", configPath)
		} else {
			// 配置文件存在，从文件加载配置
			configData, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("无法读取配置文件: %v", err)
			}

			err = json.Unmarshal(configData, &AppConfig)
			if err != nil {
				return fmt.Errorf("无法解析配置文件: %v", err)
			}

			fmt.Printf("已从 %s 加载配置\n", configPath)
		}
	} else {
		fmt.Println("使用默认配置")
	}

	// 确保上传目录存在
	if _, err := os.Stat(AppConfig.UploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(AppConfig.UploadDir, 0755)
		if err != nil {
			return fmt.Errorf("无法创建上传目录: %v", err)
		}
	}

	return nil
}
