package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// PKCS5的分组是以8为单位
// PKCS7的分组长度为1-255
func PKCS7Padding(org []byte, blockSize int) []byte {
	pad := blockSize - len(org)%blockSize
	padText := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(org, padText...)
}

// 通过AES方式解密密文
func PKCS7UnPadding(org []byte) []byte {
	l := len(org)
	pad := org[l-1]
	// org[0:4]
	return org[:l-int(pad)]
}

// AES加密
func AESEncrypt(org, key []byte) ([]byte, error) {
	// 检验秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 对明文进行补码
	org = PKCS7Padding(org, block.BlockSize())
	// 设置加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])
	// 创建密文缓冲区
	encrypted := make([]byte, len(org))
	// 加密
	blockMode.CryptBlocks(encrypted, org)
	// 返回密文
	return encrypted, nil
}

// AES解密
func AESDecrypt(cipherTxt, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	// 创建明文缓存
	org := make([]byte, len(cipherTxt))
	// 开始解密
	blockMode.CryptBlocks(org, cipherTxt)

	// 去码
	org = PKCS7UnPadding(org)
	// 返回明文
	return org, nil
}
