package utils

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func NewTaskID() string {
	return "task-" + uuid.NewV4().String()
}

func NewReqID() string {
	return "req-" + uuid.NewV4().String()
}

func ConcatURL(baseURL string, urls ...string) string {
	u, _ := url.Parse(baseURL)
	var newUrls []string
	newUrls = append(newUrls, u.Path)
	newUrls = append(newUrls, urls...)
	u.Path = path.Join(newUrls...)
	fullPath, _ := url.PathUnescape(u.String())
	return fullPath
}

func ConcatParamURL(baseURL string, params url.Values) string {
	if len(params) == 0 {
		return baseURL
	}
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func GetSnowidFromUUID(uuidValue string) uint64 {
	data := strings.Split(uuidValue, "-")
	var allBuf []byte
	for _, d := range data {
		buf, err := hex.DecodeString(d)
		if err == nil {
			allBuf = append(allBuf, buf...)
		}
	}
	tmp := make([]byte, 8)
	tmp[0] = allBuf[0]
	tmp[1] = allBuf[1]
	tmp[2] = allBuf[10]
	tmp[3] = allBuf[11]
	tmp[4] = allBuf[12]
	tmp[5] = allBuf[13]
	tmp[6] = allBuf[14]
	tmp[7] = allBuf[15]
	return binary.BigEndian.Uint64(tmp)
}

func GetUUIDFromSnowID(snowID uint64) string {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, snowID)
	uid := make([]byte, 16)
	uid[0] = tmp[0]
	uid[1] = tmp[1]
	uid[10] = tmp[2]
	uid[11] = tmp[3]
	uid[12] = tmp[4]
	uid[13] = tmp[5]
	uid[14] = tmp[6]
	uid[15] = tmp[7]
	// 设置uuid版本信息
	uid[6] = (uid[6] & 0x0f) | 0x40 // Version 4
	uid[8] = (uid[8] & 0x3f) | 0x80 // Variant is 10
	var buf [36]byte
	hex.Encode(buf[:8], uid[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uid[10:])
	return string(buf[:])
}

// RemoveDuplicated 去重
func RemoveDuplicated(codes []string) []string {
	result := make([]string, 0)
	codeMap := make(map[string]struct{})
	for _, code := range codes {
		if _, ok := codeMap[code]; !ok {
			codeMap[code] = struct{}{}
			result = append(result, code)
		}
	}
	return result
}

// SetHeader set header'value as encoded string.
func SetHeader(c *gin.Context, key, value string) {
	c.Request.Header.Set(key, url.QueryEscape(value))
}
