/*
aes加密和解密相关函数，经过封装，更好用一些
*/
package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"errors"
)

// 256位aes加密, key会用sha256进行哈希，所以key的长度没有限制，但最好要超过256位。
func AesEncrypt256_SHA256_BYTE(origData []byte, key []byte) ([]byte, error) {
	ar := sha256.Sum256(key)
	return AesEncrypt(origData, ar[:sha256.Size])
}

// 256位aes加密, key会用sha256进行哈希，所以key的长度没有限制，但最好要超过256位。
func AesEncrypt256_SHA256(origData []byte, key string) ([]byte, error) {
	ar := sha256.Sum256([]byte(key))
	return AesEncrypt(origData, ar[:sha256.Size])
}

// 256位aes解密, key会用sha256进行哈希，所以key的长度没有限制，但最好要超过256位。
func AesDecrypt256_SHA256_BYTE(crypted []byte, key []byte) ([]byte, error) {
	ar := sha256.Sum256(key)
	return AesDecrypt(crypted, ar[:sha256.Size])
}

// 256位aes解密, key会用sha256进行哈希，所以key的长度没有限制，但最好要超过256位。
func AesDecrypt256_SHA256(crypted []byte, key string) ([]byte, error) {
	ar := sha256.Sum256([]byte(key))
	return AesDecrypt(crypted, ar[:sha256.Size])
}

// 128位aes加密, key会被md5，所以长度没有限制，但最好要超过128位。
func AesEncrypt128_MD5(origData []byte, key string) ([]byte, error) {
	ar := md5.Sum([]byte(key))
	return AesEncrypt(origData, ar[:md5.Size])
}

// 128位aes解密, key会被md5。
func AesDecrypt128_MD5(crypted []byte, key string) ([]byte, error) {
	ar := md5.Sum([]byte(key))
	return AesDecrypt(crypted, ar[:md5.Size])
}

//aes加密，CBC模式，PKCS5Padding，key的长度可以为16、24或32字节
//初始化向量与key相同
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	//fmt.Printf("blockSize=%v\n", blockSize)
	origData = pKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//aes解密，key必须是16字节的整数倍
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	//	fmt.Printf("1--blockSize=%v, block:%+v, key_len: %v\n", blockSize, block, len(key))
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	//判断如果不是blockSize的整数倍，则返回错误
	if len(crypted)%blockSize != 0 {
		//		fmt.Printf("1.5-error--crtpted=%v, crypted:%+v \n", crypted, len(crypted))
		return origData, errors.New("msg.crypted len is not a multiple of th bolockSize ")
	}
	//	fmt.Printf("2--origData=%v, crypted:%+v \n", origData, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	//	fmt.Printf("3--origData=%v, crypted:%+v \n", origData, len(crypted))
	origData, err = pKCS5UnPadding(origData)
	//	fmt.Printf("4--origData=%v, er:%+v \n", origData, err)
	return origData, err
}

func zeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func zeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return origData, nil
	}
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	if length < unpadding || unpadding < 0 {
		err := errors.New("decrypt failed")
		return nil, err
	}
	return origData[:(length - unpadding)], nil
}
