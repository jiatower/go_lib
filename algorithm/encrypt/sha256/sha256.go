/*
md5加密，封装一下，更简单易用
*/
package sha256

import (
	"crypto/sha256"
	"encoding/base64"
)

//hash,传入string，先hash，然后base64
func SHA256_BASE64(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	hashed := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashed)
}

//hash,传入string，先hash，然后base64
func SHA256_BASE64_BYTE(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashed)
}
