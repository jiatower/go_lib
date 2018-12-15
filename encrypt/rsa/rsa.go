//实现用openssl生成的公私钥对（公钥采用PKCS8格式）进行加密、解密、签名、验证的功能。
//秘钥生成方法：
//openssl genrsa -out private.pem 2048
//openssl pkcs8 -topk8 -inform PEM -outform PEM -in private.pem  -out private_pkcs8.pem -nocrypt
//openssl rsa -in private.pem -pubout -out public.pem
package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

//在1024位秘钥的前提下，origData的长度不能超过117字节。
//在2048位秘钥的前提下，origData的长度不能超过245字节。
func EncryptPKCS1v15(origData []byte, publicKey []byte) ([]byte, error) {
	pub, err := parsePublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func DecryptPKCS1v15(cipherText []byte, privateKey []byte) ([]byte, error) {
	priv, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherText)
}

//在1024位秘钥的前提下，origData的长度不能超过62字节。
//在2048位秘钥的前提下，origData的长度不能超过190字节。
func EncryptOAEP_SHA256(origData []byte, publicKey []byte) ([]byte, error) {
	pub, err := parsePublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, origData, nil)
}

func DecryptOAEP_SHA256(cipherText []byte, privateKey []byte) ([]byte, error) {
	priv, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, cipherText, nil)
}

func SignPSS_SHA256(message []byte, privateKey []byte) ([]byte, error) {
	priv, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write(message)
	hashed := hash.Sum(nil)

	pssopt := new(rsa.PSSOptions)
	pssopt.Hash = crypto.SHA256
	pssopt.SaltLength = 64

	return rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hashed, pssopt)
}

//返回值为nil表示验证通过
func VerifyPSS_SHA256(message []byte, publicKey []byte, sig []byte) error {
	pub, err := parsePublicKey(publicKey)
	if err != nil {
		return err
	}
	hash := sha256.New()
	hash.Write(message)
	hashed := hash.Sum(nil)

	pssopt := new(rsa.PSSOptions)
	pssopt.Hash = crypto.SHA256
	pssopt.SaltLength = 64

	return rsa.VerifyPSS(pub, crypto.SHA256, hashed, sig, pssopt)
}

func parsePublicKey(publicKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubInterface.(*rsa.PublicKey), nil
}

func parsePrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	privInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privInterface.(*rsa.PrivateKey), nil
}
