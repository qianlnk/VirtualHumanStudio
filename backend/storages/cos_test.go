package storages

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
)

const (
	bucketAigcOps = "vhs-data"
)

func TestGetToken(t *testing.T) {
	cos, err := NewCOS(COSConfig{
		Bucket:    bucketAigcOps,
		Region:    "ap-guangzhou",
		Appid:     "1256179956",
		SecretID:  "xxxxxx",
		SecretKey: "xxxxxx",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(cos.UploadToken(context.Background(), &PutPolicy{Key: "502/avatar/aaa.jpg", Expires: 1800}))
}

func TestCOS(t *testing.T) {
	cos, err := NewCOS(COSConfig{
		Bucket:    bucketAigcOps,
		Region:    "ap-guangzhou",
		Appid:     "1256179956",
		SecretID:  "xxxxxx",
		SecretKey: "xxxxxx",
	})

	if err != nil {
		panic(err)
	}

	//fmt.Println(cos.UploadToken("", &PutPolicy{Key: "aaa.jpg", Expires: 1800}))

	key := uuid.New().String()
	fmt.Println("upload key: ", key)
	file, err := os.Open("/Users/xiezhenjia/go/src/github.com/qianlnk/VirtualHumanStudio/backend/data/1/uploads/0c81e4a0-a52f-4ec5-a92e-2076407df4fa.jpeg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = cos.SaveFile(context.Background(), key, file)
	if err != nil {
		panic(err)
	}

	info, err := cos.StatFile(context.Background(), key)
	if err != nil {
		panic(err)
	}
	fmt.Println("upload info: ", info)

	data, err := cos.FetchFile(context.Background(), key)
	if err != nil {
		panic(err)
	}

	_ = data
	// fmt.Println("==", string(data))

	url, err := cos.GetFileUrl(context.Background(), key)
	if err != nil {
		panic(err)
	}
	fmt.Println("url: ", url)

	// cpKey := uuid.New().String()
	// fmt.Println("copy key: ", cpKey)
	// err = cos.Copy(context.Background(), key, cpKey)
	// if err != nil {
	// 	panic(err)
	// }

	// info, err = cos.StatFile(context.Background(), cpKey)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("copy ingo: ", info)

	// mvKey := uuid.New().String()
	// fmt.Println("move key: ", mvKey)
	// err = cos.Move(context.Background(), key, mvKey)
	// if err != nil {
	// 	panic(err)
	// }
	// info, err = cos.StatFile(context.Background(), mvKey)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("move info: ", info)

	// err = cos.DeleteFile(context.Background(), cpKey)
	// if err != nil {
	// 	panic(err)
	// }

	// info, err = cos.StatFile(context.Background(), cpKey)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("ddelete info: ", info)
}

// func TestFetch(t *testing.T) {
// 	cos, err := NewCOS(COSConfig{
// 		Bucket:    bucketAigcOps,
// 		Region:    "ap-guangzhou",
// 		Appid:     "1256179956",
// 		SecretID:  "xxxxxx",
// 		SecretKey: "xxxxxx",
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	data, err := cos.FetchFile(context.Background(), "entity_test_3_1")
// 	if err != nil {
// 		panic(err)
// 	}

// 	scanner := bufio.NewScanner(strings.NewReader(string(data)))

// 	size := 0
// 	for scanner.Scan() {
// 		fmt.Println(scanner.Text())
// 		size++

// 		if size > 2 {
// 			break
// 		}
// 	}
// }
