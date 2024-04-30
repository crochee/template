package utils

import (
	"encoding/base64"
	"strings"

	"template/pkg/crypto"
)

var (
	// 固定前缀(勿修改)
	prefix = "V4At0ENS"
	// 密钥，长度需要为16，24或者32，此处采用16（勿修改）
	key = []byte("HzC9G9mhe3HfpuNz")
)

func AESEncrypt(plain string) (string, error) {
	if plain == "" || strings.HasPrefix(plain, prefix) {
		return plain, nil
	}
	Encode, err := crypto.AESEncrypt([]byte(plain), key)
	if err != nil {
		return "", err
	}
	encrypted := base64.StdEncoding.EncodeToString(Encode)
	return prefix + encrypted, nil
}

func AESDecrypt(encode string) (string, error) {
	if encode == "" || len(encode) <= 8 {
		return encode, nil
	}
	encoded, err := base64.StdEncoding.DecodeString(encode[8:])
	if err != nil {
		return "", err
	}

	decryptBytes, err := crypto.AESDecrypt(encoded, key)
	if err != nil {
		return "", err
	}

	return string(decryptBytes), nil
}
