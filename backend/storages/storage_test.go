package storages

// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"gitlab.xunlei.cn/database/xfile/app/utils"
// )

// var qiniu *QiniuStorage
// var ali *OSS

// func Init() {
// 	utils.InitHttpClient(utils.HttpClientConfig{ConnectTimeout: int64(time.Second) * 10})
// 	qiniu = NewQiniuStorage(QiniuConfig{
// 		AccessKey: "xxxxxx",
// 		SecretKey: "xxxxxx",
// 		Expires:   60,
// 		Zone:      2,
// 		URI:       "xfile-test.a.88cdn.com",
// 		Bucket:    "xfile",
// 	})
// 	qiniu.Init()
// 	ali = NewOSS(OSSConfig{
// 		Endpoint:        "oss-cn-shenzhen.aliyuncs.com",
// 		AccessKeyID:     "xxxxxx",
// 		AccessKeySecret: "xxxxxx",
// 		Bucket:          "account-xfile-sz",
// 		Callback: CallbackConfig{
// 			URL:      "https://account-hbz-test.lianxiangcloud.com/upload_callback",
// 			BodyType: "application/json",
// 		},
// 	})
// 	ali.Init()
// }

// func TestQiniu(t *testing.T) {
// 	Init()
// 	// fmt.Println(qiniu.StatFile("", "10030102/audit-illegal/y86E.jpg"))

// 	// fmt.Println(qiniu.Move("", "10030102/audit-illegal/y86E.jpg", "", "10030102/audit-review/y86E.jpg"))

// 	// fmt.Println(qiniu.DeleteFile("", "test"))

// 	// fmt.Println(qiniu.Copy("", "10030102/audit-illegal/y86E.jpg", "", "10030102/audit-review/y86E.jpg"))

// 	fmt.Println(qiniu.FetchFile("", "2rvk4e3gkdnl7u1kl0k/test/abc.jpg", "300x300"))
// }

// func TestAli(t *testing.T) {
// 	Init()
// 	fmt.Println(ali.StatFile("", "xfile_test.jpg"))

// 	// fmt.Println(ali.Move("", "10030102/audit-illegal/y86E.jpg", "", "10030102/audit-review/y86E.jpg1"))

// 	// fmt.Println(ali.DeleteFile("", "test"))

// 	fmt.Println(ali.Copy("", "xfile_test.jpg", "", "xfile_test.jpg/300x300"))
// 	fmt.Println(ali.StatFile("", "xfile_test.jpg/300x300"))

// 	// fmt.Println(ali.FetchFile("", "10030102/audit-illegal/y86E.jpg", "m"))
// }

// // func TestAliToken(t *testing.T) {
// // 	Init()
// // 	fmt.Println(ali.UploadToken("avatar", &PutPolicy{
// // 		Key:         "502/avatar/abcdefg",
// // 		Expires:     1800,
// // 		CallbackURL: "https://account-hbz-test.lianxiangcloud.com/upload_callback",
// // 		ReturnBody:  `{"key": ${object}, "etag": ${etag}, "size": "${size}"}`,
// // 	}))
// // }
