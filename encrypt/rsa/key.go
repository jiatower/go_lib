package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/youmark/pkcs8"
)

/*
生成RSA私钥密钥, 私钥格式为PKCS8格式
公钥为x509
bits为生成密钥位数
*/
func GenerateRsaPKCS8Key(bits int) (privateKey, publicKey []byte, e error) {
	pri, e := rsa.GenerateKey(rand.Reader, bits)
	if e != nil {
		return
	}

	derStream, e := pkcs8.ConvertPrivateKeyToPKCS8(pri)
	if e != nil {
		return
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derStream,
	}
	priKey := pem.EncodeToMemory(block)
	//数据库统一存储不包含头文件的串
	pri_key := strings.Replace(string(priKey), "-----BEGIN PRIVATE KEY-----", "", -1)
	pri_key = strings.TrimSpace(strings.Replace(pri_key, "-----END PRIVATE KEY-----", "", -1))
	privateKey = []byte(pri_key)

	// 生成公钥文件
	pub := &pri.PublicKey
	derPkix, e := x509.MarshalPKIXPublicKey(pub)
	if e != nil {
		return
	}
	block2 := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey := pem.EncodeToMemory(block2)
	//数据库统一存储不包含头文件的串
	pub_key := strings.Replace(string(pubKey), "-----BEGIN PUBLIC KEY-----", "", -1)
	pub_key = strings.TrimSpace(strings.Replace(pub_key, "-----END PUBLIC KEY-----", "", -1))
	publicKey = []byte(pub_key)
	return

}
