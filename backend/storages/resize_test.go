package storages

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestResize(t *testing.T) {
	body, err := ioutil.ReadFile("./5.webp")
	if err != nil {
		fmt.Println(1, err)
		return
	}

	data, err := ResizeImage(body, 800, 600)
	if err != nil {
		fmt.Println(2, err)
		return
	}

	out, err := os.Create("2test.png")
	if err != nil {
		fmt.Println(3, err)
		return
	}

	out.Write(data)
	out.Close()

	// time.Sleep(time.Second * 2)

	// body, err = ioutil.ReadFile("./2test.png")
	// if err != nil {
	// 	fmt.Println(1, err)
	// 	return
	// }

	// _, err = ResizeImage(body, 400, 400)
	// if err != nil {
	// 	fmt.Println(3, err)
	// 	return
	// }
}

func BenchmarkResize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		body, err := ioutil.ReadFile("./1.gif")
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = ResizeImage(body, 400, 300)
		if err != nil {
			fmt.Println(err)
			return
		}
		// out, err := os.Create(fmt.Sprintf("test%d.gif", i))
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// out.Write(data)
		// out.Close()
	}

}
