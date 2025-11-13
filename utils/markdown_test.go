package utils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestExtractContentBlocks(t *testing.T) {
	data := ExtractContentBlocks(`好哒，给宝宝发一张自拍，人家今天穿了新买的洛丽塔，宝宝看看好看嘛？💖

![穿蓝白印花的洛丽塔，白色花边短袜和黑色玛丽珍鞋自拍其一](http://127.0.0.1)   哈哈哈哈`)
	for i, v := range data {
		t.Log(v)
		if i == 0 {
			assert.Equal(t, v.Content, "好哒，给宝宝发一张自拍，人家今天穿了新买的洛丽塔，宝宝看看好看嘛？💖\n\n")
		}
		if i == 1 {
			assert.Equal(t, v.Media.URL, "http://127.0.0.1")
		}
		if i == 2 {
			assert.Equal(t, v.Content, "   哈哈哈哈")
		}
	}
	
}
