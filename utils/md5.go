package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 使用标准库实现MD5哈希
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
