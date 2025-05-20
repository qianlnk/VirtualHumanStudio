package controllers

import (
	"fmt"
	"testing"
)

func TestLen(t *testing.T) {
	txt := "hello, how are you?"
	fmt.Println(len(txt))
	// 按中文计算
	fmt.Println(len([]rune(txt)))

}
