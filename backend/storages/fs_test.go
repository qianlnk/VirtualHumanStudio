package storages

import (
	"context"
	"fmt"
	"testing"
)

func TestFS(t *testing.T) {
	fs, err := NewFS(FSConfig{
		BasePath: "./data",
	})
	if err != nil {
		panic(err)
	}

	infos, err := fs.List(context.TODO(), "fs", 0, 1000)
	if err != nil {
		panic(err)
	}

	for _, info := range infos {
		fmt.Println(info.Name)
		fmt.Println(info.Path)
	}

	// key := uuid.New().String()
	// fmt.Println("upload key: ", key)
	// text := "123123123123123123"
	// err = fs.SaveFile(context.Background(), key, strings.NewReader(text))
	// if err != nil {
	// 	panic(err)
	// }

	// info, err := fs.StatFile(context.Background(), key)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("upload info: ", info)

	// data, err := fs.FetchFile(context.Background(), key)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("==", string(data))
}
