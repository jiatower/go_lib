/*
md5加密，封装一下，更简单易用
*/
package md5

import (
	"crypto/md5"
	"encoding/hex"
)

//MD5加密
func MD5Sum(key string) string {
	ar := md5.Sum([]byte(key))
	return hex.EncodeToString(ar[:md5.Size])
}
