package utils

import (
	"strings"
)

/*
字符串截取函数

参数：
	str:带截取字符串
	begin:开始截取位置
	length:截取长度
*/
func SubString(str string, begin, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)
	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	// 返回子串
	return string(rs[begin:end])
}

//生成公共串
func GenKey(sep string, keys ...interface{}) (key string) {
	arr := make([]string, 0, len(keys))
	for _, v := range keys {
		arr = append(arr, ToString(v))
	}
	key = strings.Join(arr, sep)
	return
}

// ContainAnySubstr: 输入string和一个slice的substring，给出这个string是否包含任何一个substring
func ContainAnySubstr(str string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}