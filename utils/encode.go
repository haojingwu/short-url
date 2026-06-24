package utils

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

// Base62字符集(0-9a-zA-Z)
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateShortCode 根据原始URL生成短链码(MD5 + Base62)
// 特点: 同一URL生成同一个code(等幂性)
func GenerateShortCode(originalURL string) string {
	// 计算MD5
	hash := md5.Sum([]byte(originalURL))

	// 取前4字节转成uint64
	// 得到的数字范围0 ~ 2^32 -1
	val := binary.BigEndian.Uint32(hash[:4])

	// 转成Base62
	return encodeBase62(uint64(val))
}

// encodeBase62将uint64编码为base62字符串
func encodeBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	for num > 0 {
		remainder := num % 62
		num = num / 62
		result = append([]byte{base62Chars[remainder]}, result...)
	}

	return string(result)
}

// GenerateShortCodeWithRetry 生成短链码,如果冲突则重试
// dbChecker:检查code是否已存在的函数
// 最多重试3次,每次在URL后面追加随机字符
func GenerateShortCodeWithRetry(originalURL string, dbChecker func(code string) bool) string {
	//第一次尝试:使用原始URL生成
	code := GenerateShortCode(originalURL)
	if !dbChecker(code) {
		return code //没有冲突 直接返回
	}

	//冲突了,重试,最多3次
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		//在URL后面追加随机字符,破坏MD5结果
		salt := fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(1000))
		modifiedURL := originalURL + salt
		code = GenerateShortCode(modifiedURL)

		if !dbChecker(code) {
			return code //重试成功
		}
	}

	//3次都冲突,用时间戳兜底
	return fmt.Sprintf("tmp%d", time.Now().UnixNano()%100000)
}

// DecodeBase62ToUint64 将Base62字符串解码为 uint64
func DecodeBase62ToUint64(str string) uint64 {
	var result uint64
	for _, char := range str {
		var val uint64
		if char >= '0' && char <= '9' {
			val = uint64(char - '0')
		} else if char >= 'a' && char <= 'z' {
			val = uint64(char - 'a' + 10)
		} else if char >= 'A' && char <= 'Z' {
			val = uint64(char - 'A' + 36)
		} else {
			continue //忽略非法字符
		}
		result = result*62 + val
	}
	return result
}
